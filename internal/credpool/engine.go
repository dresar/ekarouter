package credpool

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var ErrAllCredentialsExhausted = errors.New("all pool credentials exhausted")
var ErrPoolPaused = errors.New("credential pool is paused")
var ErrNoMembers = errors.New("pool has no members")
var ErrManualMemberRequired = errors.New("manual strategy requires explicit member_id")

type Engine struct {
	store   *Store
	mu      sync.RWMutex
	cursors map[string]*uint64
}

func NewEngine(store *Store) *Engine {
	return &Engine{
		store:   store,
		cursors: make(map[string]*uint64),
	}
}

func (e *Engine) Select(ctx context.Context, req *SelectRequest) (*Member, error) {
	start := time.Now()

	pool, err := e.store.GetPool(ctx, req.PoolID)
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	if pool.Status == PoolPaused {
		return nil, ErrPoolPaused
	}

	policy, err := e.store.GetPolicy(ctx, req.PoolID)
	if err != nil {
		return nil, fmt.Errorf("select policy: %w", err)
	}

	if policy.DisableAutoRotation && policy.Strategy != StrategyManual {
		policy.Strategy = StrategyManual
	}

	members, err := e.store.ListMembers(ctx, req.PoolID)
	if err != nil {
		return nil, fmt.Errorf("select members: %w", err)
	}
	if len(members) == 0 {
		return nil, ErrNoMembers
	}

	var eligible []Member
	for _, m := range members {
		if m.IsAvailable() {
			eligible = append(eligible, m)
		}
	}
	if len(eligible) == 0 {
		return nil, ErrAllCredentialsExhausted
	}

	var selected *Member
	var reason string

	switch policy.Strategy {
	case StrategySingle:
		selected, reason = e.selectSingle(eligible)
	case StrategyPriorityFallback:
		selected, reason = e.selectPriorityFallback(eligible)
	case StrategyRoundRobin:
		selected, reason = e.selectRoundRobin(req.PoolID, eligible)
	case StrategyWeightedRoundRobin:
		selected, reason = e.selectWeightedRoundRobin(eligible)
	case StrategyLRU:
		selected, reason = e.selectLRU(eligible)
	case StrategyQuotaAware:
		selected, reason = e.selectQuotaAware(eligible)
	case StrategyLowestFailureRate:
		selected, reason = e.selectLowestFailureRate(eligible)
	case StrategyEnvironmentAware:
		selected, reason, err = e.selectEnvironmentAware(req.Environment, pool.Environment, eligible)
		if err != nil {
			return nil, err
		}
	case StrategyManual:
		selected, reason, err = e.selectManual(req.MemberID, eligible)
		if err != nil {
			return nil, err
		}
	default:
		selected, reason = e.selectPriorityFallback(eligible)
	}

	if selected == nil {
		return nil, ErrAllCredentialsExhausted
	}

	elapsed := time.Since(start)
	_ = e.store.RecordRotationDecision(ctx, &RotationDecision{
		PoolID:           req.PoolID,
		SelectedMemberID: selected.ID,
		Strategy:         policy.Strategy,
		Reason:           reason,
		CandidatesCount:  len(eligible),
		RequestID:        req.RequestID,
		LatencyUs:        int(elapsed.Microseconds()),
	})

	return selected, nil
}

func (e *Engine) selectSingle(eligible []Member) (*Member, string) {
	if len(eligible) == 0 {
		return nil, ""
	}
	return &eligible[0], "single_active_member"
}

func (e *Engine) selectPriorityFallback(eligible []Member) (*Member, string) {
	if len(eligible) == 0 {
		return nil, ""
	}
	return &eligible[0], fmt.Sprintf("priority_fallback_p%d", eligible[0].Priority)
}

func (e *Engine) selectRoundRobin(poolID string, eligible []Member) (*Member, string) {
	if len(eligible) == 0 {
		return nil, ""
	}

	e.mu.Lock()
	cursor, ok := e.cursors[poolID]
	if !ok {
		var c uint64
		cursor = &c
		e.cursors[poolID] = cursor
	}
	e.mu.Unlock()

	idx := atomic.AddUint64(cursor, 1) % uint64(len(eligible))
	selected := &eligible[idx]
	return selected, fmt.Sprintf("round_robin_idx_%d", idx)
}

func (e *Engine) selectWeightedRoundRobin(eligible []Member) (*Member, string) {
	if len(eligible) == 0 {
		return nil, ""
	}

	totalWeight := 0
	for _, m := range eligible {
		w := m.Weight
		if w <= 0 {
			w = 1
		}
		totalWeight += w
	}

	r := rand.Intn(totalWeight)
	cumulative := 0
	for i := range eligible {
		w := eligible[i].Weight
		if w <= 0 {
			w = 1
		}
		cumulative += w
		if r < cumulative {
			return &eligible[i], fmt.Sprintf("weighted_rr_w%d", eligible[i].Weight)
		}
	}

	return &eligible[0], "weighted_rr_fallback"
}

func (e *Engine) selectLRU(eligible []Member) (*Member, string) {
	if len(eligible) == 0 {
		return nil, ""
	}

	var oldest *Member
	for i := range eligible {
		m := &eligible[i]
		if m.LastUsedAt == nil {
			return m, "least_recently_used_never_used"
		}
		if oldest == nil || (oldest.LastUsedAt != nil && m.LastUsedAt.Before(*oldest.LastUsedAt)) {
			oldest = m
		}
	}
	return oldest, "least_recently_used"
}

func (e *Engine) selectEnvironmentAware(reqEnv, poolEnv string, eligible []Member) (*Member, string, error) {
	if len(eligible) == 0 {
		return nil, "", ErrAllCredentialsExhausted
	}
	if reqEnv != "" && poolEnv != "" && !strings.EqualFold(reqEnv, poolEnv) {
		return &eligible[0], fmt.Sprintf("environment_mismatch_fallback_%s_vs_%s", reqEnv, poolEnv), nil
	}
	return &eligible[0], fmt.Sprintf("environment_matched_%s", poolEnv), nil
}

func (e *Engine) selectQuotaAware(eligible []Member) (*Member, string) {
	if len(eligible) == 0 {
		return nil, ""
	}

	best := &eligible[0]
	bestScore := e.quotaScore(best)
	for i := 1; i < len(eligible); i++ {
		score := e.quotaScore(&eligible[i])
		if score > bestScore {
			best = &eligible[i]
			bestScore = score
		}
	}
	return best, fmt.Sprintf("quota_aware_score_%d", bestScore)
}

func (e *Engine) quotaScore(m *Member) int {
	score := 1000
	score -= m.TotalRequests
	if m.FailureCount > 0 {
		score -= m.FailureCount * 10
	}
	return score
}

func (e *Engine) selectLowestFailureRate(eligible []Member) (*Member, string) {
	if len(eligible) == 0 {
		return nil, ""
	}

	best := &eligible[0]
	bestRate := failureRate(best)
	for i := 1; i < len(eligible); i++ {
		rate := failureRate(&eligible[i])
		if rate < bestRate {
			best = &eligible[i]
			bestRate = rate
		}
	}
	return best, fmt.Sprintf("lowest_failure_rate_%.2f", bestRate)
}

func failureRate(m *Member) float64 {
	if m.TotalRequests == 0 {
		return 0
	}
	return float64(m.FailureCount) / float64(m.TotalRequests)
}

func (e *Engine) selectManual(memberID string, eligible []Member) (*Member, string, error) {
	if memberID == "" {
		return nil, "", ErrManualMemberRequired
	}
	for i := range eligible {
		if eligible[i].ID == memberID {
			return &eligible[i], "manual_selection", nil
		}
	}
	return nil, "", fmt.Errorf("requested member %s not found or not available", memberID)
}

func (e *Engine) MarkSuccess(ctx context.Context, memberID string) error {
	return e.store.MarkMemberUsed(ctx, memberID, true)
}

func (e *Engine) MarkFailure(ctx context.Context, memberID string, policy *RotationPolicy) error {
	if err := e.store.MarkMemberUsed(ctx, memberID, false); err != nil {
		return err
	}

	cooldownDur := time.Duration(policy.CooldownSeconds) * time.Second
	if cooldownDur > 0 {
		until := time.Now().Add(cooldownDur)
		return e.store.SetMemberCooldown(ctx, memberID, until)
	}
	return nil
}
