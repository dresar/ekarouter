package rotator

import (
	"errors"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dresar/ekarouter/internal/vault"
)

type Strategy string

const (
	StrategyPriority        Strategy = "priority"
	StrategyRoundRobin      Strategy = "round_robin"
	StrategyRandom          Strategy = "random"
	StrategyLeastUsed       Strategy = "least_used"
	StrategyLowestErrorRate Strategy = "lowest_error_rate"
	StrategyHealthBased     Strategy = "health_based"
)

type Rotator struct {
	mu      sync.Mutex
	cursors map[string]*uint64
}

func NewRotator() *Rotator {
	return &Rotator{
		cursors: make(map[string]*uint64),
	}
}

func (r *Rotator) getCursor(key string) *uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cursors == nil {
		r.cursors = make(map[string]*uint64)
	}
	c, ok := r.cursors[key]
	if !ok {
		var val uint64
		c = &val
		r.cursors[key] = c
	}
	return c
}

func (r *Rotator) Select(candidates []*vault.Credential, strat Strategy) (*vault.Credential, error) {
	key := "default"
	for _, c := range candidates {
		if c.ProviderID != "" {
			key = c.ProviderID
			break
		}
	}
	return r.SelectForPool(key, candidates, strat)
}

func (r *Rotator) SelectForPool(poolKey string, candidates []*vault.Credential, strat Strategy) (*vault.Credential, error) {
	now := time.Now().UTC()
	var eligible []*vault.Credential

	for _, c := range candidates {
		if c.Status != "active" {
			continue
		}
		if c.ExpiresAt != nil && now.After(*c.ExpiresAt) {
			continue
		}
		if c.CooldownUntil != nil && c.CooldownUntil.After(now) {
			continue
		}
		eligible = append(eligible, c)
	}

	if len(eligible) == 0 {
		return nil, errors.New("no eligible credential available")
	}

	healthy := make([]*vault.Credential, 0, len(eligible))
	for _, c := range eligible {
		if c.HealthState != vault.HealthUnhealthy && c.HealthState != vault.HealthExpired {
			healthy = append(healthy, c)
		}
	}
	if len(healthy) > 0 {
		eligible = healthy
	}

	switch strat {
	case StrategyRoundRobin:
		cursor := r.getCursor(poolKey)
		idx := atomic.AddUint64(cursor, 1) - 1
		return eligible[idx%uint64(len(eligible))], nil

	case StrategyRandom:
		return eligible[rand.Intn(len(eligible))], nil

	case StrategyLeastUsed:
		sort.SliceStable(eligible, func(i, j int) bool {
			return eligible[i].RequestCount < eligible[j].RequestCount
		})
		return eligible[0], nil

	case StrategyLowestErrorRate:
		sort.SliceStable(eligible, func(i, j int) bool {
			rateI := calcErrorRate(eligible[i])
			rateJ := calcErrorRate(eligible[j])
			return rateI < rateJ
		})
		return eligible[0], nil

	case StrategyHealthBased:
		sort.SliceStable(eligible, func(i, j int) bool {
			scoreI := healthScore(eligible[i])
			scoreJ := healthScore(eligible[j])
			return scoreI > scoreJ
		})
		return eligible[0], nil

	case StrategyPriority:
		fallthrough
	default:
		sort.SliceStable(eligible, func(i, j int) bool {
			if eligible[i].Priority != eligible[j].Priority {
				return eligible[i].Priority < eligible[j].Priority
			}
			return eligible[i].RequestCount < eligible[j].RequestCount
		})
		return eligible[0], nil
	}
}

func calcErrorRate(c *vault.Credential) float64 {
	if c.RequestCount == 0 {
		return 0.0
	}
	return float64(c.ErrorCount) / float64(c.RequestCount)
}

func healthScore(c *vault.Credential) int {
	score := 100
	switch c.HealthState {
	case vault.HealthHealthy:
		score += 50
	case vault.HealthDegraded:
		score -= 30
	case vault.HealthUnhealthy:
		score -= 80
	}
	if c.LastError != "" {
		score -= 20
	}
	score -= int(calcErrorRate(c) * 50)
	return score
}
