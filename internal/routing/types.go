package routing

import (
	"time"
)

type Strategy string

const (
	StrategyPriority   Strategy = "priority"
	StrategyRoundRobin Strategy = "round_robin"
)

type RouteItem struct {
	ID           string
	RouteID      string
	ProviderID   string
	ProviderKind string
	AccountID    string
	ModelName    string
	Priority     int
	Weight       int
	Enabled      bool
	TimeoutMs    int
	MaxRetries   int
}

type Route struct {
	ID        string
	Name      string
	Strategy  Strategy
	Enabled   bool
	Items     []RouteItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Account struct {
	ID         string
	ProviderID string
	Name       string
	AuthType   string
	State      string
	Priority   int
	Enabled    bool
	ExpiresAt  *time.Time
}

func (a *Account) IsActive() bool {
	if !a.Enabled || a.State != "active" {
		return false
	}
	if a.ExpiresAt != nil && time.Now().After(*a.ExpiresAt) {
		return false
	}
	return true
}

type Target struct {
	ProviderID   string
	ProviderKind string
	AccountID    string
	ModelName    string
	Timeout      time.Duration
	MaxRetries   int
}
