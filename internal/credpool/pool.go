package credpool

import (
	"time"
)

type Strategy string

const (
	StrategySingle             Strategy = "single"
	StrategyPriorityFallback   Strategy = "priority_fallback"
	StrategyRoundRobin         Strategy = "round_robin"
	StrategyWeightedRoundRobin Strategy = "weighted_round_robin"
	StrategyLRU                Strategy = "least_recently_used"
	StrategyQuotaAware         Strategy = "quota_aware"
	StrategyLowestFailureRate  Strategy = "lowest_failure_rate"
	StrategyEnvironmentAware   Strategy = "environment_aware"
	StrategyManual             Strategy = "manual"
)

type MemberStatus string

const (
	StatusActive              MemberStatus = "active"
	StatusDisabled            MemberStatus = "disabled"
	StatusPendingValidation   MemberStatus = "pending_validation"
	StatusValid               MemberStatus = "valid"
	StatusInvalid             MemberStatus = "invalid"
	StatusExpired             MemberStatus = "expired"
	StatusRevoked             MemberStatus = "revoked"
	StatusPermissionDenied    MemberStatus = "permission_denied"
	StatusRateLimited         MemberStatus = "rate_limited"
	StatusQuotaExhausted      MemberStatus = "quota_exhausted"
	StatusCoolingDown         MemberStatus = "cooling_down"
	StatusProviderUnavailable MemberStatus = "provider_unavailable"
	StatusRefreshRequired     MemberStatus = "refresh_required"
	StatusUnknown             MemberStatus = "unknown"
)

type HealthStatus string

const (
	HealthValid                  HealthStatus = "valid"
	HealthExpired                HealthStatus = "expired"
	HealthRevoked                HealthStatus = "revoked"
	HealthInsufficientPermission HealthStatus = "insufficient_permission"
	HealthRateLimited            HealthStatus = "rate_limited"
	HealthProviderUnavailable    HealthStatus = "provider_unavailable"
	HealthInvalidConfig          HealthStatus = "invalid_config"
	HealthUnsupportedValidation  HealthStatus = "unsupported_validation"
	HealthUnknown                HealthStatus = "unknown"
)

type PoolStatus string

const (
	PoolActive   PoolStatus = "active"
	PoolPaused   PoolStatus = "paused"
	PoolDisabled PoolStatus = "disabled"
)

type Pool struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	OwnerID     string     `json:"owner_id"`
	ProviderID  string     `json:"provider_id,omitempty"`
	Environment string     `json:"environment"`
	Status      PoolStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Member struct {
	ID            string       `json:"id"`
	PoolID        string       `json:"pool_id"`
	AccountID     string       `json:"account_id"`
	Priority      int          `json:"priority"`
	Weight        int          `json:"weight"`
	Status        MemberStatus `json:"status"`
	CooldownUntil *time.Time   `json:"cooldown_until,omitempty"`
	LastUsedAt    *time.Time   `json:"last_used_at,omitempty"`
	LastSuccessAt *time.Time   `json:"last_success_at,omitempty"`
	LastFailureAt *time.Time   `json:"last_failure_at,omitempty"`
	ExpiresAt     *time.Time   `json:"expires_at,omitempty"`
	FailureCount  int          `json:"failure_count"`
	SuccessCount  int          `json:"success_count"`
	TotalRequests int          `json:"total_requests"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

func (m *Member) IsAvailable() bool {
	if m.ExpiresAt != nil && time.Now().After(*m.ExpiresAt) {
		return false
	}
	if m.CooldownUntil != nil && time.Now().Before(*m.CooldownUntil) {
		return false
	}
	if m.Status == StatusActive || m.Status == StatusValid {
		return true
	}
	if m.Status == StatusCoolingDown && m.CooldownUntil != nil && time.Now().After(*m.CooldownUntil) {
		return true
	}
	return false
}

type RotationPolicy struct {
	ID                  string    `json:"id"`
	PoolID              string    `json:"pool_id"`
	Strategy            Strategy  `json:"strategy"`
	DisableAutoRotation bool      `json:"disable_auto_rotation"`
	MaxConcurrent       int       `json:"max_concurrent"`
	CooldownSeconds     int       `json:"cooldown_seconds"`
	QuotaAware          bool      `json:"quota_aware"`
	RetryTransientOnly  bool      `json:"retry_transient_only"`
	MaxRetries          int       `json:"max_retries"`
	FallbackOnPermanent bool      `json:"fallback_on_permanent"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type HealthResult struct {
	MemberID  string       `json:"member_id"`
	Status    HealthStatus `json:"status"`
	LatencyMs int          `json:"latency_ms"`
	Message   string       `json:"message"`
	CheckedAt time.Time    `json:"checked_at"`
}

type EventType string

const (
	EventCreated         EventType = "created"
	EventStatusChanged   EventType = "status_changed"
	EventRotated         EventType = "rotated"
	EventHealthChecked   EventType = "health_checked"
	EventExpired         EventType = "expired"
	EventRefreshed       EventType = "refreshed"
	EventCooldownStarted EventType = "cooldown_started"
	EventCooldownEnded   EventType = "cooldown_ended"
	EventPaused          EventType = "paused"
	EventResumed         EventType = "resumed"
)

type Event struct {
	ID        int64     `json:"id"`
	PoolID    string    `json:"pool_id"`
	MemberID  string    `json:"member_id"`
	AccountID string    `json:"account_id"`
	Type      EventType `json:"event_type"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

type RotationDecision struct {
	ID               int64     `json:"id"`
	PoolID           string    `json:"pool_id"`
	SelectedMemberID string    `json:"selected_member_id"`
	Strategy         Strategy  `json:"strategy"`
	Reason           string    `json:"reason"`
	CandidatesCount  int       `json:"candidates_count"`
	RequestID        string    `json:"request_id"`
	LatencyUs        int       `json:"latency_us"`
	CreatedAt        time.Time `json:"created_at"`
}

type SelectRequest struct {
	PoolID      string `json:"pool_id"`
	RequestID   string `json:"request_id"`
	Environment string `json:"environment"`
	MemberID    string `json:"member_id"`
}
