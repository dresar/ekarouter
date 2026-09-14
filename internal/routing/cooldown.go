package routing

import (
	"sync"
	"time"
)

type entry struct {
	failures      int
	cooldownUntil time.Time
}

type CooldownManager struct {
	mu      sync.RWMutex
	entries map[string]*entry
}

func NewCooldownManager() *CooldownManager {
	return &CooldownManager{
		entries: make(map[string]*entry),
	}
}

func (c *CooldownManager) MarkSuccess(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, id)
}

func (c *CooldownManager) MarkFailure(id string, baseCooldown time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.entries[id]
	if !ok {
		e = &entry{}
		c.entries[id] = e
	}
	e.failures++

	multiplier := 1 << (e.failures - 1)
	if multiplier > 16 {
		multiplier = 16
	}

	cd := baseCooldown * time.Duration(multiplier)
	e.cooldownUntil = time.Now().Add(cd)
}

func (c *CooldownManager) IsCoolingDown(id string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[id]
	if !ok {
		return false
	}
	return time.Now().Before(e.cooldownUntil)
}

func (c *CooldownManager) Reset(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, id)
}
