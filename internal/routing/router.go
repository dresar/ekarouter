package routing

import (
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Router struct {
	mu        sync.RWMutex
	aliases   map[string]string
	routes    map[string]*Route
	accounts  map[string]*Account
	cursors   map[string]*uint64
	cooldowns *CooldownManager
}

func NewRouter(cooldowns *CooldownManager) *Router {
	if cooldowns == nil {
		cooldowns = NewCooldownManager()
	}
	return &Router{
		aliases:   make(map[string]string),
		routes:    make(map[string]*Route),
		accounts:  make(map[string]*Account),
		cursors:   make(map[string]*uint64),
		cooldowns: cooldowns,
	}
}

func (r *Router) SetAlias(alias, target string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.aliases[alias] = target
}

func (r *Router) ResolveModel(model string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if target, ok := r.aliases[model]; ok {
		return target
	}
	return model
}

func (r *Router) SetRoute(route *Route) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[route.Name] = route
	if _, ok := r.cursors[route.Name]; !ok {
		var c uint64
		r.cursors[route.Name] = &c
	}
}

func (r *Router) SetAccount(acc *Account) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accounts[acc.ID] = acc
}

func (r *Router) SelectTargets(requestedModel string) ([]Target, error) {
	targetName := r.ResolveModel(requestedModel)

	r.mu.RLock()
	route, hasRoute := r.routes[targetName]
	r.mu.RUnlock()

	if !hasRoute || !route.Enabled || len(route.Items) == 0 {
		return nil, errors.New("no eligible route or combo for target: " + targetName)
	}

	r.mu.RLock()
	var eligible []RouteItem
	for _, item := range route.Items {
		if !item.Enabled {
			continue
		}
		if item.AccountID != "" {
			acc, exists := r.accounts[item.AccountID]
			if !exists || !acc.IsActive() {
				continue
			}
			if r.cooldowns.IsCoolingDown(item.AccountID) {
				continue
			}
		}
		eligible = append(eligible, item)
	}
	cursorPtr := r.cursors[route.Name]
	r.mu.RUnlock()

	if len(eligible) == 0 {
		return nil, errors.New("all route candidates are cooling down or disabled")
	}

	sort.SliceStable(eligible, func(i, j int) bool {
		return eligible[i].Priority < eligible[j].Priority
	})

	if route.Strategy == StrategyRoundRobin && cursorPtr != nil && len(eligible) > 1 {
		idx := atomic.AddUint64(cursorPtr, 1) % uint64(len(eligible))
		rotated := make([]RouteItem, len(eligible))
		copy(rotated, eligible[idx:])
		copy(rotated[len(eligible)-int(idx):], eligible[:idx])
		eligible = rotated
	}

	var targets []Target
	for _, item := range eligible {
		timeout := 60 * time.Second
		if item.TimeoutMs > 0 {
			timeout = time.Duration(item.TimeoutMs) * time.Millisecond
		}
		retries := 2
		if item.MaxRetries > 0 {
			retries = item.MaxRetries
		}
		targets = append(targets, Target{
			ProviderID:   item.ProviderID,
			ProviderKind: item.ProviderKind,
			AccountID:    item.AccountID,
			ModelName:    item.ModelName,
			Timeout:      timeout,
			MaxRetries:   retries,
		})
	}

	return targets, nil
}
