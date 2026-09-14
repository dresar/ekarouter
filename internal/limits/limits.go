package limits

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"
)

type QuotaState string

const (
	StateUnknown   QuotaState = "unknown"
	StateAvailable QuotaState = "available"
	StateLimited   QuotaState = "limited"
	StateExhausted QuotaState = "exhausted"
	StateSuspended QuotaState = "suspended"
	StateInvalid   QuotaState = "invalid"
	StateExpired   QuotaState = "expired"
	StateDisabled  QuotaState = "disabled"
)

type QuotaRecord struct {
	ID            string     `json:"id"`
	ReferenceType string     `json:"reference_type"`
	ReferenceID   string     `json:"reference_id"`
	Metric        string     `json:"metric"`
	UsedValue     int64      `json:"used_value"`
	MaxValue      int64      `json:"max_value"`
	Period        string     `json:"period"`
	ResetAt       *time.Time `json:"reset_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Engine struct {
	db       *sql.DB
	mu       sync.RWMutex
	memRates map[string][]time.Time
}

func NewEngine(db *sql.DB) *Engine {
	return &Engine{
		db:       db,
		memRates: make(map[string][]time.Time),
	}
}

func (e *Engine) AllowRate(key string, limit int, window time.Duration) bool {
	if limit <= 0 {
		return true
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	timestamps := e.memRates[key]
	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= limit {
		e.memRates[key] = valid
		return false
	}

	valid = append(valid, now)
	e.memRates[key] = valid
	return true
}

func (e *Engine) SetQuota(ctx context.Context, refType, refID, metric, period string, maxVal int64, resetAt *time.Time) error {
	now := time.Now().UTC()
	id := uuid.New().String()

	query := `
INSERT INTO quota_records (id, reference_type, reference_id, metric, used_value, max_value, period, reset_at, updated_at)
VALUES (?, ?, ?, ?, 0, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET max_value = excluded.max_value, reset_at = excluded.reset_at, updated_at = excluded.updated_at`

	var existingID string
	err := e.db.QueryRowContext(ctx, "SELECT id FROM quota_records WHERE reference_type = ? AND reference_id = ? AND metric = ?", refType, refID, metric).Scan(&existingID)
	if err == nil && existingID != "" {
		_, err = e.db.ExecContext(ctx, "UPDATE quota_records SET max_value = ?, reset_at = ?, updated_at = ? WHERE id = ?", maxVal, resetAt, now, existingID)
		return err
	}

	_, err = e.db.ExecContext(ctx, query, id, refType, refID, metric, maxVal, period, resetAt, now)
	return err
}

func (e *Engine) CheckQuota(ctx context.Context, refType, refID, metric string) (QuotaState, int64, int64, error) {
	var usedVal, maxVal int64
	var resetAt sql.NullTime

	err := e.db.QueryRowContext(ctx, `
SELECT used_value, max_value, reset_at
FROM quota_records
WHERE reference_type = ? AND reference_id = ? AND metric = ?`, refType, refID, metric).Scan(&usedVal, &maxVal, &resetAt)

	if errorsIsNoRows(err) {
		return StateUnknown, 0, 0, nil
	}
	if err != nil {
		return StateUnknown, 0, 0, err
	}

	if resetAt.Valid && time.Now().UTC().After(resetAt.Time) {
		_, _ = e.db.ExecContext(ctx, "UPDATE quota_records SET used_value = 0, updated_at = ? WHERE reference_type = ? AND reference_id = ? AND metric = ?", time.Now().UTC(), refType, refID, metric)
		usedVal = 0
	}

	if maxVal > 0 && usedVal >= maxVal {
		return StateExhausted, usedVal, maxVal, nil
	}
	if maxVal > 0 && float64(usedVal)/float64(maxVal) >= 0.9 {
		return StateLimited, usedVal, maxVal, nil
	}
	return StateAvailable, usedVal, maxVal, nil
}

func (e *Engine) IncrementQuota(ctx context.Context, refType, refID, metric string, delta int64) error {
	now := time.Now().UTC()
	_, err := e.db.ExecContext(ctx, `
UPDATE quota_records
SET used_value = used_value + ?, updated_at = ?
WHERE reference_type = ? AND reference_id = ? AND metric = ?`, delta, now, refType, refID, metric)
	return err
}

func (e *Engine) SetCooldown(ctx context.Context, credID string, duration time.Duration, reason string) error {
	now := time.Now().UTC()
	endsAt := now.Add(duration)
	cdID := uuid.New().String()

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE vault_credentials SET cooldown_until = ?, last_error = ?, updated_at = ? WHERE id = ?", endsAt, reason, now, credID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO credential_cooldowns (id, credential_id, reason, retry_after_seconds, started_at, ends_at)
VALUES (?, ?, ?, ?, ?, ?)`, cdID, credID, reason, int(duration.Seconds()), now, endsAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (e *Engine) IsInCooldown(ctx context.Context, credID string) (bool, time.Duration, error) {
	var cdUntil sql.NullTime
	err := e.db.QueryRowContext(ctx, "SELECT cooldown_until FROM vault_credentials WHERE id = ?", credID).Scan(&cdUntil)
	if err != nil {
		return false, 0, err
	}
	if !cdUntil.Valid {
		return false, 0, nil
	}

	now := time.Now().UTC()
	if cdUntil.Time.After(now) {
		return true, cdUntil.Time.Sub(now), nil
	}
	return false, 0, nil
}

func errorsIsNoRows(err error) bool {
	return err == sql.ErrNoRows
}
