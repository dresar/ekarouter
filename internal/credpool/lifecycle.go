package credpool

import (
	"context"
	"sync"
	"time"
)

type LifecycleManager struct {
	store      *Store
	checker    *HealthChecker
	mu         sync.Mutex
	refreshing map[string]bool
}

func NewLifecycleManager(store *Store, checker *HealthChecker) *LifecycleManager {
	return &LifecycleManager{
		store:      store,
		checker:    checker,
		refreshing: make(map[string]bool),
	}
}

func (lm *LifecycleManager) DetectExpired(ctx context.Context, poolID string) ([]Member, error) {
	members, err := lm.store.ListMembers(ctx, poolID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var expired []Member
	for _, m := range members {
		if m.ExpiresAt != nil && now.After(*m.ExpiresAt) && m.Status != StatusExpired {
			_ = lm.store.UpdateMemberStatus(ctx, m.ID, StatusExpired)
			_ = lm.store.RecordEvent(ctx, &Event{
				PoolID:    poolID,
				MemberID:  m.ID,
				AccountID: m.AccountID,
				Type:      EventExpired,
				Details:   "credential expired at " + m.ExpiresAt.Format(time.RFC3339),
			})
			m.Status = StatusExpired
			expired = append(expired, m)
		}
	}
	return expired, nil
}

func (lm *LifecycleManager) PauseUnusable(ctx context.Context, poolID string) error {
	members, err := lm.store.ListMembers(ctx, poolID)
	if err != nil {
		return err
	}

	for _, m := range members {
		if shouldPause(m) {
			_ = lm.store.UpdateMemberStatus(ctx, m.ID, StatusDisabled)
			_ = lm.store.RecordEvent(ctx, &Event{
				PoolID:    poolID,
				MemberID:  m.ID,
				AccountID: m.AccountID,
				Type:      EventPaused,
				Details:   "paused due to status: " + string(m.Status),
			})
		}
	}
	return nil
}

func shouldPause(m Member) bool {
	switch m.Status {
	case StatusRevoked, StatusPermissionDenied, StatusInvalid:
		return true
	default:
		return false
	}
}

func (lm *LifecycleManager) RestoreAfterValidation(ctx context.Context, poolID string, memberID string) error {
	if lm.checker == nil {
		return nil
	}

	member, err := lm.store.GetMember(ctx, memberID)
	if err != nil {
		return err
	}

	result := lm.checker.CheckMember(ctx, member, poolID)
	if result.Status == HealthValid {
		_ = lm.store.UpdateMemberStatus(ctx, memberID, StatusActive)
		_ = lm.store.ClearMemberCooldown(ctx, memberID)
		_ = lm.store.RecordEvent(ctx, &Event{
			PoolID:    poolID,
			MemberID:  memberID,
			AccountID: member.AccountID,
			Type:      EventResumed,
			Details:   "restored after successful validation",
		})
	}
	return nil
}

func (lm *LifecycleManager) TryRefresh(ctx context.Context, memberID string) bool {
	lm.mu.Lock()
	if lm.refreshing[memberID] {
		lm.mu.Unlock()
		return false
	}
	lm.refreshing[memberID] = true
	lm.mu.Unlock()

	defer func() {
		lm.mu.Lock()
		delete(lm.refreshing, memberID)
		lm.mu.Unlock()
	}()

	_ = lm.store.UpdateMemberStatus(ctx, memberID, StatusRefreshRequired)
	return true
}

func (lm *LifecycleManager) ClearCooldowns(ctx context.Context, poolID string) error {
	members, err := lm.store.ListMembers(ctx, poolID)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, m := range members {
		if m.CooldownUntil != nil && now.After(*m.CooldownUntil) && m.Status == StatusCoolingDown {
			_ = lm.store.ClearMemberCooldown(ctx, m.ID)
			_ = lm.store.RecordEvent(ctx, &Event{
				PoolID:    poolID,
				MemberID:  m.ID,
				AccountID: m.AccountID,
				Type:      EventCooldownEnded,
				Details:   "cooldown period ended",
			})
		}
	}
	return nil
}
