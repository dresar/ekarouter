package credpool

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreatePool(ctx context.Context, p *Pool) error {
	if p.ID == "" {
		p.ID = "pool_" + uuid.NewString()[:8]
	}
	if p.Environment == "" {
		p.Environment = "production"
	}
	if p.Status == "" {
		p.Status = PoolActive
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, `
INSERT INTO credential_pools (id, name, owner_id, provider_id, environment, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.OwnerID, p.ProviderID, p.Environment, p.Status, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
INSERT INTO rotation_policies (id, pool_id, strategy) VALUES (?, ?, ?)`,
		"rp_"+uuid.NewString()[:8], p.ID, StrategyPriorityFallback)
	if err != nil {
		return fmt.Errorf("create default rotation policy: %w", err)
	}

	return nil
}

func (s *Store) GetPool(ctx context.Context, id string) (*Pool, error) {
	var p Pool
	var providerID sql.NullString
	err := s.db.QueryRowContext(ctx, `
SELECT id, name, owner_id, provider_id, environment, status, created_at, updated_at
FROM credential_pools WHERE id = ?`, id).Scan(
		&p.ID, &p.Name, &p.OwnerID, &providerID, &p.Environment, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get pool %s: %w", id, err)
	}
	if providerID.Valid {
		p.ProviderID = providerID.String
	}
	return &p, nil
}

func (s *Store) ListPools(ctx context.Context, ownerID string) ([]Pool, error) {
	query := "SELECT id, name, owner_id, provider_id, environment, status, created_at, updated_at FROM credential_pools"
	var args []any
	if ownerID != "" {
		query += " WHERE owner_id = ?"
		args = append(args, ownerID)
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list pools: %w", err)
	}
	defer rows.Close()

	var pools []Pool
	for rows.Next() {
		var p Pool
		var providerID sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.OwnerID, &providerID, &p.Environment, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		if providerID.Valid {
			p.ProviderID = providerID.String
		}
		pools = append(pools, p)
	}
	return pools, nil
}

func (s *Store) UpdatePoolStatus(ctx context.Context, id string, status PoolStatus) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE credential_pools SET status = ?, updated_at = ? WHERE id = ?",
		status, time.Now(), id)
	return err
}

func (s *Store) DeletePool(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM credential_pools WHERE id = ?", id)
	return err
}

func (s *Store) AddMember(ctx context.Context, m *Member) error {
	if m.ID == "" {
		m.ID = "pm_" + uuid.NewString()[:8]
	}
	if m.Status == "" {
		m.Status = StatusActive
	}
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, `
INSERT INTO pool_members (id, pool_id, account_id, priority, weight, status, cooldown_until,
    last_used_at, last_success_at, last_failure_at, expires_at, failure_count, success_count,
    total_requests, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.PoolID, m.AccountID, m.Priority, m.Weight, m.Status,
		m.CooldownUntil, m.LastUsedAt, m.LastSuccessAt, m.LastFailureAt, m.ExpiresAt,
		m.FailureCount, m.SuccessCount, m.TotalRequests, m.CreatedAt, m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (s *Store) ListMembers(ctx context.Context, poolID string) ([]Member, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, pool_id, account_id, priority, weight, status, cooldown_until,
    last_used_at, last_success_at, last_failure_at, expires_at,
    failure_count, success_count, total_requests, created_at, updated_at
FROM pool_members WHERE pool_id = ? ORDER BY priority ASC, weight DESC`, poolID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		var cooldown, lastUsed, lastSuccess, lastFailure, expires sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.PoolID, &m.AccountID, &m.Priority, &m.Weight, &m.Status,
			&cooldown, &lastUsed, &lastSuccess, &lastFailure, &expires,
			&m.FailureCount, &m.SuccessCount, &m.TotalRequests, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			continue
		}
		if cooldown.Valid {
			m.CooldownUntil = &cooldown.Time
		}
		if lastUsed.Valid {
			m.LastUsedAt = &lastUsed.Time
		}
		if lastSuccess.Valid {
			m.LastSuccessAt = &lastSuccess.Time
		}
		if lastFailure.Valid {
			m.LastFailureAt = &lastFailure.Time
		}
		if expires.Valid {
			m.ExpiresAt = &expires.Time
		}
		members = append(members, m)
	}
	return members, nil
}

func (s *Store) GetMember(ctx context.Context, id string) (*Member, error) {
	var m Member
	var cooldown, lastUsed, lastSuccess, lastFailure, expires sql.NullTime
	err := s.db.QueryRowContext(ctx, `
SELECT id, pool_id, account_id, priority, weight, status, cooldown_until,
    last_used_at, last_success_at, last_failure_at, expires_at,
    failure_count, success_count, total_requests, created_at, updated_at
FROM pool_members WHERE id = ?`, id).Scan(
		&m.ID, &m.PoolID, &m.AccountID, &m.Priority, &m.Weight, &m.Status,
		&cooldown, &lastUsed, &lastSuccess, &lastFailure, &expires,
		&m.FailureCount, &m.SuccessCount, &m.TotalRequests, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get member %s: %w", id, err)
	}
	if cooldown.Valid {
		m.CooldownUntil = &cooldown.Time
	}
	if lastUsed.Valid {
		m.LastUsedAt = &lastUsed.Time
	}
	if lastSuccess.Valid {
		m.LastSuccessAt = &lastSuccess.Time
	}
	if lastFailure.Valid {
		m.LastFailureAt = &lastFailure.Time
	}
	if expires.Valid {
		m.ExpiresAt = &expires.Time
	}
	return &m, nil
}

func (s *Store) RemoveMember(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM pool_members WHERE id = ?", id)
	return err
}

func (s *Store) UpdateMemberStatus(ctx context.Context, id string, status MemberStatus) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE pool_members SET status = ?, updated_at = ? WHERE id = ?",
		status, time.Now(), id)
	return err
}

func (s *Store) MarkMemberUsed(ctx context.Context, id string, success bool) error {
	now := time.Now()
	if success {
		_, err := s.db.ExecContext(ctx, `
UPDATE pool_members SET
    last_used_at = ?, last_success_at = ?, success_count = success_count + 1,
    total_requests = total_requests + 1, updated_at = ?
WHERE id = ?`, now, now, now, id)
		return err
	}
	_, err := s.db.ExecContext(ctx, `
UPDATE pool_members SET
    last_used_at = ?, last_failure_at = ?, failure_count = failure_count + 1,
    total_requests = total_requests + 1, updated_at = ?
WHERE id = ?`, now, now, now, id)
	return err
}

func (s *Store) SetMemberCooldown(ctx context.Context, id string, until time.Time) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE pool_members SET cooldown_until = ?, status = ?, updated_at = ? WHERE id = ?",
		until, StatusCoolingDown, time.Now(), id)
	return err
}

func (s *Store) ClearMemberCooldown(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE pool_members SET cooldown_until = NULL, status = ?, updated_at = ? WHERE id = ?",
		StatusActive, time.Now(), id)
	return err
}

func (s *Store) GetPolicy(ctx context.Context, poolID string) (*RotationPolicy, error) {
	var rp RotationPolicy
	err := s.db.QueryRowContext(ctx, `
SELECT id, pool_id, strategy, disable_auto_rotation, max_concurrent, cooldown_seconds,
    quota_aware, retry_transient_only, max_retries, fallback_on_permanent, created_at, updated_at
FROM rotation_policies WHERE pool_id = ?`, poolID).Scan(
		&rp.ID, &rp.PoolID, &rp.Strategy, &rp.DisableAutoRotation, &rp.MaxConcurrent,
		&rp.CooldownSeconds, &rp.QuotaAware, &rp.RetryTransientOnly, &rp.MaxRetries,
		&rp.FallbackOnPermanent, &rp.CreatedAt, &rp.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get policy for pool %s: %w", poolID, err)
	}
	return &rp, nil
}

func (s *Store) SetPolicy(ctx context.Context, rp *RotationPolicy) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO rotation_policies (id, pool_id, strategy, disable_auto_rotation, max_concurrent,
    cooldown_seconds, quota_aware, retry_transient_only, max_retries, fallback_on_permanent, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(pool_id) DO UPDATE SET
    strategy = excluded.strategy,
    disable_auto_rotation = excluded.disable_auto_rotation,
    max_concurrent = excluded.max_concurrent,
    cooldown_seconds = excluded.cooldown_seconds,
    quota_aware = excluded.quota_aware,
    retry_transient_only = excluded.retry_transient_only,
    max_retries = excluded.max_retries,
    fallback_on_permanent = excluded.fallback_on_permanent,
    updated_at = excluded.updated_at`,
		rp.ID, rp.PoolID, rp.Strategy, rp.DisableAutoRotation, rp.MaxConcurrent,
		rp.CooldownSeconds, rp.QuotaAware, rp.RetryTransientOnly, rp.MaxRetries,
		rp.FallbackOnPermanent, time.Now())
	return err
}

func (s *Store) RecordEvent(ctx context.Context, e *Event) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO credential_events (pool_id, member_id, account_id, event_type, details)
VALUES (?, ?, ?, ?, ?)`, e.PoolID, e.MemberID, e.AccountID, e.Type, e.Details)
	return err
}

func (s *Store) RecordHealthCheck(ctx context.Context, hr *HealthResult, poolID string) error {
	id := "hc_" + uuid.NewString()[:8]
	_, err := s.db.ExecContext(ctx, `
INSERT INTO credential_health_checks (id, member_id, pool_id, status, latency_ms, message, checked_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`, id, hr.MemberID, poolID, hr.Status, hr.LatencyMs, hr.Message, hr.CheckedAt)
	return err
}

func (s *Store) RecordRotationDecision(ctx context.Context, rd *RotationDecision) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO rotation_decisions (pool_id, selected_member_id, strategy, reason, candidates_count, request_id, latency_us)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rd.PoolID, rd.SelectedMemberID, rd.Strategy, rd.Reason, rd.CandidatesCount, rd.RequestID, rd.LatencyUs)
	return err
}

func (s *Store) ListEvents(ctx context.Context, poolID string, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, pool_id, member_id, account_id, event_type, details, created_at
FROM credential_events WHERE pool_id = ? ORDER BY created_at DESC LIMIT ?`, poolID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var memberID, accountID sql.NullString
		if err := rows.Scan(&e.ID, &e.PoolID, &memberID, &accountID, &e.Type, &e.Details, &e.CreatedAt); err != nil {
			continue
		}
		if memberID.Valid {
			e.MemberID = memberID.String
		}
		if accountID.Valid {
			e.AccountID = accountID.String
		}
		events = append(events, e)
	}
	return events, nil
}
