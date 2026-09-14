package rbac

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	Role         string    `json:"role"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OwnerID     string    `json:"owner_id"`
	Environment string    `json:"environment"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Environment struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ProjectID   string    `json:"project_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type ClientToken struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	TokenHash  string     `json:"-"`
	UserID     string     `json:"user_id,omitempty"`
	ProjectID  string     `json:"project_id,omitempty"`
	Scopes     string     `json:"scopes"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	hash := pbkdf2Sha256([]byte(password), salt, 100000, 32)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", 100000, hex.EncodeToString(salt), hex.EncodeToString(hash)), nil
}

func CheckPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}
	expectedHash, err := hex.DecodeString(parts[3])
	if err != nil {
		return false
	}
	actualHash := pbkdf2Sha256([]byte(password), salt, 100000, len(expectedHash))
	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1
}

func pbkdf2Sha256(password, salt []byte, iter, keyLen int) []byte {
	prf := func(data []byte) []byte {
		h := sha256.New()
		h.Write(password)
		h.Write(data)
		return h.Sum(nil)
	}

	numBlocks := (keyLen + sha256.Size - 1) / sha256.Size
	var result []byte

	for block := 1; block <= numBlocks; block++ {
		var blockBytes [4]byte
		blockBytes[0] = byte(block >> 24)
		blockBytes[1] = byte(block >> 16)
		blockBytes[2] = byte(block >> 8)
		blockBytes[3] = byte(block)

		u := prf(append(salt, blockBytes[:]...))
		t := make([]byte, len(u))
		copy(t, u)

		for i := 1; i < iter; i++ {
			u = prf(u)
			for j := 0; j < len(t); j++ {
				t[j] ^= u[j]
			}
		}
		result = append(result, t...)
	}

	return result[:keyLen]
}

func (s *Service) CreateUser(ctx context.Context, email, password, displayName, role string) (*User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}
	if role == "" {
		role = "developer"
	}

	pwdHash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	u := &User{
		ID:           uuid.New().String(),
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: pwdHash,
		DisplayName:  displayName,
		Role:         role,
		Enabled:      true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	query := `
INSERT INTO users (id, email, password_hash, display_name, role, enabled, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 1, ?, ?)`

	_, err = s.db.ExecContext(ctx, query, u.ID, u.Email, u.PasswordHash, u.DisplayName, u.Role, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
	query := `
SELECT id, email, password_hash, display_name, role, enabled, created_at, updated_at
FROM users WHERE email = ?`

	var u User
	var enInt int
	err := s.db.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(email))).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &enInt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	u.Enabled = enInt == 1
	if !u.Enabled {
		return nil, errors.New("user account disabled")
	}

	if !CheckPassword(password, u.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}
	return &u, nil
}

func (s *Service) CheckPermission(role, requiredPermission string) bool {
	if role == "admin" {
		return true
	}
	if requiredPermission == "" {
		return true
	}

	rolePermissions := map[string][]string{
		"developer": {
			"providers.read", "credentials.read", "credentials.create",
			"credentials.update", "credentials.test", "tools.read", "tools.execute",
			"projects.read", "usage.read",
		},
		"viewer": {
			"providers.read", "tools.read", "projects.read", "usage.read",
		},
	}

	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == requiredPermission || p == "*" {
			return true
		}
	}
	return false
}

func (s *Service) CreateProject(ctx context.Context, name, ownerID, env, desc string) (*Project, error) {
	if name == "" {
		return nil, errors.New("project name is required")
	}
	if env == "" {
		env = "development"
	}
	p := &Project{
		ID:          uuid.New().String(),
		Name:        name,
		OwnerID:     ownerID,
		Environment: env,
		Description: desc,
		Enabled:     true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	query := `
INSERT INTO projects (id, name, owner_id, environment, description, enabled, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 1, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, p.ID, p.Name, p.OwnerID, p.Environment, p.Description, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ListProjects(ctx context.Context) ([]*Project, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, owner_id, environment, description, enabled, created_at, updated_at FROM projects ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Project
	for rows.Next() {
		var p Project
		var enInt int
		if err := rows.Scan(&p.ID, &p.Name, &p.OwnerID, &p.Environment, &p.Description, &enInt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Enabled = enInt == 1
		list = append(list, &p)
	}
	return list, rows.Err()
}

func (s *Service) CreateClientToken(ctx context.Context, name, userID, projectID, scopes string, expiresDays int) (string, *ClientToken, error) {
	bytes := make([]byte, 24)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", nil, err
	}
	rawToken := "eka_pat_" + hex.EncodeToString(bytes)

	hashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	var exp *time.Time
	if expiresDays > 0 {
		t := time.Now().UTC().Add(time.Duration(expiresDays) * 24 * time.Hour)
		exp = &t
	}

	ct := &ClientToken{
		ID:        uuid.New().String(),
		Name:      name,
		TokenHash: tokenHash,
		UserID:    userID,
		ProjectID: projectID,
		Scopes:    scopes,
		ExpiresAt: exp,
		CreatedAt: time.Now().UTC(),
	}

	var uID, pID any
	if userID != "" {
		uID = userID
	}
	if projectID != "" {
		pID = projectID
	}

	query := `
INSERT INTO client_tokens (id, name, token_hash, user_id, project_id, scopes, expires_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, ct.ID, ct.Name, ct.TokenHash, uID, pID, ct.Scopes, ct.ExpiresAt, ct.CreatedAt)
	if err != nil {
		return "", nil, err
	}

	return rawToken, ct, nil
}

func (s *Service) VerifyClientToken(ctx context.Context, rawToken string) (*ClientToken, error) {
	hashBytes := sha256.Sum256([]byte(strings.TrimSpace(rawToken)))
	tokenHash := hex.EncodeToString(hashBytes[:])

	query := `
SELECT id, name, token_hash, user_id, project_id, scopes, expires_at, created_at, last_used_at
FROM client_tokens WHERE token_hash = ?`

	var ct ClientToken
	var uID, pID sql.NullString
	var exp, lastUsed sql.NullTime

	err := s.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&ct.ID, &ct.Name, &ct.TokenHash, &uID, &pID, &ct.Scopes, &exp, &ct.CreatedAt, &lastUsed,
	)
	if err != nil {
		return nil, errors.New("invalid client token")
	}

	if exp.Valid && time.Now().UTC().After(exp.Time) {
		return nil, errors.New("client token expired")
	}

	if uID.Valid {
		ct.UserID = uID.String
	}
	if pID.Valid {
		ct.ProjectID = pID.String
	}
	if exp.Valid {
		t := exp.Time
		ct.ExpiresAt = &t
	}

	now := time.Now().UTC()
	_, _ = s.db.ExecContext(ctx, "UPDATE client_tokens SET last_used_at = ? WHERE id = ?", now, ct.ID)

	return &ct, nil
}
