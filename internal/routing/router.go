package routing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Router struct {
	mu            sync.RWMutex
	aliases       map[string]string
	routes        map[string]*Route
	accounts      map[string]*Account
	providers     map[string]string
	modelProvider map[string]string
	cursors       map[string]*uint64
	cooldowns     *CooldownManager
}

func NewRouter(cooldowns *CooldownManager) *Router {
	if cooldowns == nil {
		cooldowns = NewCooldownManager()
	}
	return &Router{
		aliases:       make(map[string]string),
		routes:        make(map[string]*Route),
		accounts:      make(map[string]*Account),
		providers:     make(map[string]string),
		modelProvider: make(map[string]string),
		cursors:       make(map[string]*uint64),
		cooldowns:     cooldowns,
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

func (r *Router) getCursor(name string) *uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.cursors[name]
	if !ok {
		var val uint64
		c = &val
		r.cursors[name] = c
	}
	return c
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

func (r *Router) SetProvider(id, kind string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[id] = kind
}

func (r *Router) SetModel(modelName, providerKind string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modelProvider[modelName] = providerKind
}

func (r *Router) LoadFromDB(ctx context.Context, database *sql.DB) error {
	provRows, err := database.QueryContext(ctx, "SELECT id, kind FROM providers WHERE enabled = 1")
	if err != nil {
		return err
	}
	defer provRows.Close()

	providersMap := make(map[string]string)
	for provRows.Next() {
		var id, kind string
		if err := provRows.Scan(&id, &kind); err == nil {
			providersMap[id] = kind
		}
	}

	modelRows, err := database.QueryContext(ctx, `
SELECT m.id, m.external_name, p.kind
FROM models m
JOIN providers p ON p.id = m.provider_id
WHERE m.enabled = 1 AND p.enabled = 1`)
	if err == nil {
		defer modelRows.Close()
	}

	modelsMap := make(map[string]string)
	if err == nil {
		for modelRows.Next() {
			var id, extName, kind string
			if err := modelRows.Scan(&id, &extName, &kind); err == nil {
				modelsMap[id] = kind
				modelsMap[extName] = kind
			}
		}
	}

	accRows, err := database.QueryContext(ctx, "SELECT id, provider_id, name, auth_type, state, priority, enabled, expires_at FROM accounts WHERE enabled = 1")
	if err != nil {
		return err
	}
	defer accRows.Close()

	accountsMap := make(map[string]*Account)
	for accRows.Next() {
		var acc Account
		var enabledInt int
		var expiresAt sql.NullTime
		if err := accRows.Scan(&acc.ID, &acc.ProviderID, &acc.Name, &acc.AuthType, &acc.State, &acc.Priority, &enabledInt, &expiresAt); err == nil {
			acc.Enabled = enabledInt == 1
			if expiresAt.Valid {
				acc.ExpiresAt = &expiresAt.Time
			}
			accountsMap[acc.ID] = &acc
		}
	}

	routeRows, err := database.QueryContext(ctx, "SELECT id, name, strategy, enabled FROM routes WHERE enabled = 1")
	if err != nil {
		return err
	}
	defer routeRows.Close()

	routesMap := make(map[string]*Route)
	for routeRows.Next() {
		var route Route
		var enabledInt int
		if err := routeRows.Scan(&route.ID, &route.Name, &route.Strategy, &enabledInt); err == nil {
			route.Enabled = enabledInt == 1

			itemRows, err := database.QueryContext(ctx, `
SELECT ri.id, ri.route_id, ri.provider_id, p.kind, COALESCE(ri.account_id, ''), COALESCE(m.external_name, ''), ri.priority, ri.weight, ri.enabled, ri.timeout_ms, ri.max_retries
FROM route_items ri
JOIN providers p ON p.id = ri.provider_id
LEFT JOIN models m ON m.id = ri.model_id
WHERE ri.route_id = ? AND ri.enabled = 1`, route.ID)

			if err == nil {
				for itemRows.Next() {
					var item RouteItem
					var itemEnabled int
					if err := itemRows.Scan(&item.ID, &item.RouteID, &item.ProviderID, &item.ProviderKind, &item.AccountID, &item.ModelName, &item.Priority, &item.Weight, &itemEnabled, &item.TimeoutMs, &item.MaxRetries); err == nil {
						item.Enabled = itemEnabled == 1
						route.Items = append(route.Items, item)
					}
				}
				itemRows.Close()
			}

			routesMap[route.Name] = &route
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = providersMap
	r.modelProvider = modelsMap
	r.accounts = accountsMap
	r.routes = routesMap
	for name := range routesMap {
		if _, ok := r.cursors[name]; !ok {
			var c uint64
			r.cursors[name] = &c
		}
	}

	return nil
}

func (r *Router) SelectTargets(requestedModel string) ([]Target, error) {
	targetName := r.ResolveModel(requestedModel)

	r.mu.RLock()
	route, hasRoute := r.routes[targetName]
	r.mu.RUnlock()

	if hasRoute && route.Enabled && len(route.Items) > 0 {
		return r.selectFromRoute(route)
	}

	return r.selectDynamicTargets(targetName)
}

func (r *Router) selectFromRoute(route *Route) ([]Target, error) {
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
		} else {
			hasActive := false
			for _, acc := range r.accounts {
				if acc.ProviderID == item.ProviderID && acc.IsActive() && !r.cooldowns.IsCoolingDown(acc.ID) {
					hasActive = true
					break
				}
			}
			if !hasActive {
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
		if eligible[i].Priority != eligible[j].Priority {
			return eligible[i].Priority < eligible[j].Priority
		}
		return eligible[i].ID < eligible[j].ID
	})

	if route.Strategy == StrategyRoundRobin && cursorPtr != nil && len(eligible) > 1 {
		idx := atomic.AddUint64(cursorPtr, 1) % uint64(len(eligible))
		rotated := make([]RouteItem, len(eligible))
		copy(rotated, eligible[idx:])
		copy(rotated[len(eligible)-int(idx):], eligible[:idx])
		eligible = rotated
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

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

		if item.AccountID != "" {
			targets = append(targets, Target{
				ProviderID:   item.ProviderID,
				ProviderKind: item.ProviderKind,
				AccountID:    item.AccountID,
				ModelName:    item.ModelName,
				Timeout:      timeout,
				MaxRetries:   retries,
			})
		} else {
			var matching []*Account
			for _, acc := range r.accounts {
				if acc.ProviderID == item.ProviderID && acc.IsActive() && !r.cooldowns.IsCoolingDown(acc.ID) {
					matching = append(matching, acc)
				}
			}
			sort.SliceStable(matching, func(i, j int) bool {
				if matching[i].Priority != matching[j].Priority {
					return matching[i].Priority < matching[j].Priority
				}
				return matching[i].ID < matching[j].ID
			})
			if len(matching) > 1 && route.Strategy == StrategyRoundRobin {
				topPriority := matching[0].Priority
				topCount := 0
				for _, acc := range matching {
					if acc.Priority == topPriority {
						topCount++
					} else {
						break
					}
				}
				if topCount > 1 {
					cursor := r.getCursor("wild:" + route.Name + ":" + item.ProviderID)
					idx := atomic.AddUint64(cursor, 1) % uint64(topCount)
					rotatedTop := make([]*Account, topCount)
					copy(rotatedTop, matching[idx:topCount])
					copy(rotatedTop[topCount-int(idx):], matching[:idx])
					copy(matching[:topCount], rotatedTop)
				}
			}
			for _, acc := range matching {
				targets = append(targets, Target{
					ProviderID:   item.ProviderID,
					ProviderKind: item.ProviderKind,
					AccountID:    acc.ID,
					ModelName:    item.ModelName,
					Timeout:      timeout,
					MaxRetries:   retries,
				})
			}
		}
	}

	return targets, nil
}

func (r *Router) selectDynamicTargets(targetName string) ([]Target, error) {
	providerKind := ""
	actualModel := targetName
	specificProviderID := ""

	if strings.Contains(targetName, "/") {
		parts := strings.SplitN(targetName, "/", 2)
		prefix := strings.ToLower(parts[0])
		r.mu.RLock()
		if kind, ok := r.providers[prefix]; ok {
			providerKind = kind
			specificProviderID = prefix
			actualModel = parts[1]
		}
		r.mu.RUnlock()
		if providerKind == "" {
			switch prefix {
			case "openai", "anthropic", "gemini", "custom":
				providerKind = prefix
				actualModel = parts[1]
			}
		}
	}

	if providerKind == "" {
		r.mu.RLock()
		if kind, ok := r.modelProvider[targetName]; ok {
			providerKind = kind
		}
		r.mu.RUnlock()
	}

	if providerKind == "" {
		lower := strings.ToLower(targetName)
		switch {
		case strings.HasPrefix(lower, "gpt-") || strings.HasPrefix(lower, "o1") || strings.HasPrefix(lower, "o3") || strings.HasPrefix(lower, "chatgpt") || strings.HasPrefix(lower, "text-embedding"):
			providerKind = "openai"
		case strings.HasPrefix(lower, "claude"):
			providerKind = "anthropic"
		case strings.HasPrefix(lower, "gemini"):
			providerKind = "gemini"
		}
	}

	if providerKind == "" {
		return nil, errors.New("no eligible route or combo for target: " + targetName)
	}

	r.mu.RLock()
	var eligible []*Account
	for _, acc := range r.accounts {
		if !acc.IsActive() {
			continue
		}
		if r.cooldowns.IsCoolingDown(acc.ID) {
			continue
		}
		if specificProviderID != "" {
			if acc.ProviderID == specificProviderID {
				eligible = append(eligible, acc)
			}
			continue
		}
		kind := r.providers[acc.ProviderID]
		if kind == "" {
			kind = acc.ProviderID
		}
		if kind == providerKind {
			eligible = append(eligible, acc)
		}
	}
	r.mu.RUnlock()

	if len(eligible) == 0 {
		return nil, fmt.Errorf("no active accounts for provider kind: %s (target: %s)", providerKind, targetName)
	}

	sort.SliceStable(eligible, func(i, j int) bool {
		if eligible[i].Priority != eligible[j].Priority {
			return eligible[i].Priority < eligible[j].Priority
		}
		return eligible[i].ID < eligible[j].ID
	})

	if len(eligible) > 1 {
		topPriority := eligible[0].Priority
		topCount := 0
		for _, acc := range eligible {
			if acc.Priority == topPriority {
				topCount++
			} else {
				break
			}
		}
		if topCount > 1 {
			cursor := r.getCursor("dyn:" + providerKind)
			idx := atomic.AddUint64(cursor, 1) % uint64(topCount)
			rotatedTop := make([]*Account, topCount)
			copy(rotatedTop, eligible[idx:topCount])
			copy(rotatedTop[topCount-int(idx):], eligible[:idx])
			copy(eligible[:topCount], rotatedTop)
		}
	}

	var targets []Target
	for _, acc := range eligible {
		targets = append(targets, Target{
			ProviderID:   acc.ProviderID,
			ProviderKind: providerKind,
			AccountID:    acc.ID,
			ModelName:    actualModel,
			Timeout:      60 * time.Second,
			MaxRetries:   2,
		})
	}
	return targets, nil
}
