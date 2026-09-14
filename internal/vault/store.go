package vault

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	db    *sql.DB
	vault *Vault
}

func NewStore(db *sql.DB, vault *Vault) *Store {
	return &Store{
		db:    db,
		vault: vault,
	}
}

func (s *Store) CreateCredential(ctx context.Context, cred *Credential, rawSecret string) error {
	if cred.Name == "" || cred.ProviderID == "" || rawSecret == "" {
		return errors.New("name, provider_id, and secret value are required")
	}
	if cred.ID == "" {
		cred.ID = uuid.New().String()
	}
	if cred.Environment == "" {
		cred.Environment = "production"
	}
	if cred.Status == "" {
		cred.Status = "active"
	}
	if cred.HealthState == "" {
		cred.HealthState = HealthUnknown
	}
	if cred.Priority <= 0 {
		cred.Priority = 10
	}

	encVal, err := s.vault.Encrypt(rawSecret)
	if err != nil {
		return fmt.Errorf("encrypt credential: %w", err)
	}
	cred.EncryptedValue = encVal
	cred.MaskedValue = MaskCredential(rawSecret)

	now := time.Now().UTC()
	cred.CreatedAt = now
	cred.UpdatedAt = now

	var projID, teamID any
	if cred.ProjectID != "" {
		projID = cred.ProjectID
	}
	if cred.TeamID != "" {
		teamID = cred.TeamID
	}

	query := `
INSERT INTO vault_credentials (
    id, name, credential_type, provider_id, project_id, team_id,
    environment, encrypted_value, masked_value, priority, tags, status,
    health_state, cooldown_until, expires_at, notes, last_used_at,
    last_validated_at, last_error, request_count, error_count, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = s.db.ExecContext(ctx, query,
		cred.ID, cred.Name, string(cred.CredentialType), cred.ProviderID,
		projID, teamID, cred.Environment, cred.EncryptedValue,
		cred.MaskedValue, cred.Priority, cred.Tags, cred.Status,
		string(cred.HealthState), cred.CooldownUntil, cred.ExpiresAt, cred.Notes,
		cred.LastUsedAt, cred.LastValidated, cred.LastError, cred.RequestCount,
		cred.ErrorCount, cred.CreatedAt, cred.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert credential: %w", err)
	}

	verID := uuid.New().String()
	_, _ = s.db.ExecContext(ctx, `
INSERT INTO credential_versions (id, credential_id, version_num, encrypted_value, created_at)
VALUES (?, ?, 1, ?, ?)`, verID, cred.ID, encVal, now)

	return nil
}

func (s *Store) GetCredential(ctx context.Context, id string) (*Credential, error) {
	query := `
SELECT id, name, credential_type, provider_id, project_id, team_id,
       environment, encrypted_value, masked_value, priority, tags, status,
       health_state, cooldown_until, expires_at, notes, last_used_at,
       last_validated_at, last_error, request_count, error_count, created_at, updated_at
FROM vault_credentials WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, id)
	return s.scanCredential(row)
}

func (s *Store) GetDecryptedSecret(ctx context.Context, id string) (string, error) {
	var encVal string
	err := s.db.QueryRowContext(ctx, "SELECT encrypted_value FROM vault_credentials WHERE id = ?", id).Scan(&encVal)
	if err != nil {
		return "", err
	}
	return s.vault.Decrypt(encVal)
}

func (s *Store) ListCredentials(ctx context.Context, providerID, projectID, environment string) ([]*Credential, error) {
	var sb strings.Builder
	sb.WriteString(`
SELECT id, name, credential_type, provider_id, project_id, team_id,
       environment, encrypted_value, masked_value, priority, tags, status,
       health_state, cooldown_until, expires_at, notes, last_used_at,
       last_validated_at, last_error, request_count, error_count, created_at, updated_at
FROM vault_credentials WHERE 1=1`)

	var args []any
	if providerID != "" {
		sb.WriteString(" AND provider_id = ?")
		args = append(args, providerID)
	}
	if projectID != "" {
		sb.WriteString(" AND project_id = ?")
		args = append(args, projectID)
	}
	if environment != "" {
		sb.WriteString(" AND environment = ?")
		args = append(args, environment)
	}
	sb.WriteString(" ORDER BY priority ASC, created_at DESC")

	rows, err := s.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Credential
	for rows.Next() {
		c, err := s.scanCredential(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (s *Store) UpdateCredential(ctx context.Context, id string, name string, priority int, tags string, notes string) error {
	now := time.Now().UTC()
	query := `
UPDATE vault_credentials
SET name = ?, priority = ?, tags = ?, notes = ?, updated_at = ?
WHERE id = ?`
	res, err := s.db.ExecContext(ctx, query, name, priority, tags, notes, now, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("credential not found")
	}
	return nil
}

func (s *Store) RotateSecret(ctx context.Context, id string, newSecret string) error {
	if newSecret == "" {
		return errors.New("new secret cannot be empty")
	}
	encVal, err := s.vault.Encrypt(newSecret)
	if err != nil {
		return fmt.Errorf("encrypt new secret: %w", err)
	}
	masked := MaskCredential(newSecret)
	now := time.Now().UTC()

	var maxVer int
	_ = s.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version_num), 0) FROM credential_versions WHERE credential_id = ?", id).Scan(&maxVer)
	nextVer := maxVer + 1

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
UPDATE vault_credentials
SET encrypted_value = ?, masked_value = ?, health_state = ?, cooldown_until = NULL, last_error = '', updated_at = ?
WHERE id = ?`, encVal, masked, string(HealthUnknown), now, id)
	if err != nil {
		return err
	}

	verID := uuid.New().String()
	_, err = tx.ExecContext(ctx, `
INSERT INTO credential_versions (id, credential_id, version_num, encrypted_value, created_at)
VALUES (?, ?, ?, ?, ?)`, verID, id, nextVer, encVal, now)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) SetStatus(ctx context.Context, id string, status string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, "UPDATE vault_credentials SET status = ?, updated_at = ? WHERE id = ?", status, now, id)
	return err
}

func (s *Store) SetHealthState(ctx context.Context, id string, state HealthState, lastErr string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `
UPDATE vault_credentials
SET health_state = ?, last_error = ?, last_validated_at = ?, updated_at = ?
WHERE id = ?`, string(state), lastErr, now, now, id)
	return err
}

func (s *Store) SetCooldown(ctx context.Context, id string, cooldownDuration time.Duration, reason string) error {
	now := time.Now().UTC()
	until := now.Add(cooldownDuration)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
UPDATE vault_credentials
SET cooldown_until = ?, last_error = ?, updated_at = ?
WHERE id = ?`, until, reason, now, id)
	if err != nil {
		return err
	}

	cdID := uuid.New().String()
	_, err = tx.ExecContext(ctx, `
INSERT INTO credential_cooldowns (id, credential_id, reason, retry_after_seconds, started_at, ends_at)
VALUES (?, ?, ?, ?, ?, ?)`, cdID, id, reason, int(cooldownDuration.Seconds()), now, until)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) RecordUsage(ctx context.Context, id string, isError bool) error {
	now := time.Now().UTC()
	query := `
UPDATE vault_credentials
SET request_count = request_count + 1,
    error_count = error_count + CASE WHEN ? THEN 1 ELSE 0 END,
    last_used_at = ?,
    updated_at = ?
WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, isError, now, now, id)
	return err
}

func (s *Store) DeleteCredential(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM vault_credentials WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("credential not found")
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func (s *Store) scanCredential(r scannable) (*Credential, error) {
	c := &Credential{}
	var credType, healthStr string
	var projID, teamID sql.NullString
	var cooldownUntil, expiresAt, lastUsed, lastVal sql.NullTime

	err := r.Scan(
		&c.ID, &c.Name, &credType, &c.ProviderID, &projID, &teamID,
		&c.Environment, &c.EncryptedValue, &c.MaskedValue, &c.Priority, &c.Tags, &c.Status,
		&healthStr, &cooldownUntil, &expiresAt, &c.Notes, &lastUsed,
		&lastVal, &c.LastError, &c.RequestCount, &c.ErrorCount, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	c.CredentialType = CredentialType(credType)
	c.HealthState = HealthState(healthStr)
	if projID.Valid {
		c.ProjectID = projID.String
	}
	if teamID.Valid {
		c.TeamID = teamID.String
	}
	if cooldownUntil.Valid {
		t := cooldownUntil.Time
		c.CooldownUntil = &t
	}
	if expiresAt.Valid {
		t := expiresAt.Time
		c.ExpiresAt = &t
	}
	if lastUsed.Valid {
		t := lastUsed.Time
		c.LastUsedAt = &t
	}
	if lastVal.Valid {
		t := lastVal.Time
		c.LastValidated = &t
	}

	return c, nil
}
