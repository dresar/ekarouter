package oauth

import (
	"sync"
	"testing"
	"time"
)

func TestOAuthStateGenerationAndConsumption(t *testing.T) {
	mgr := NewManager()

	rawState, codeChallenge, err := mgr.GenerateState("provider-google", 5*time.Minute)
	if err != nil {
		t.Fatalf("generate state: %v", err)
	}

	if rawState == "" || codeChallenge == "" {
		t.Fatal("empty state or challenge")
	}

	rec, err := mgr.ConsumeState(rawState)
	if err != nil {
		t.Fatalf("consume state: %v", err)
	}

	if rec.ProviderID != "provider-google" {
		t.Errorf("expected provider-google, got %s", rec.ProviderID)
	}

	_, err = mgr.ConsumeState(rawState)
	if err == nil {
		t.Error("expected second consume to fail (single-use)")
	}
}

func TestOAuthExpiredState(t *testing.T) {
	mgr := NewManager()

	rawState, _, err := mgr.GenerateState("p1", -1*time.Minute)
	if err != nil {
		t.Fatalf("generate state: %v", err)
	}

	_, err = mgr.ConsumeState(rawState)
	if err == nil {
		t.Error("expected error for expired state")
	}
}

func TestAntiStampedeRefreshLock(t *testing.T) {
	mgr := NewManager()

	var wg sync.WaitGroup
	var activeCount int
	var maxActive int
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			unlock := mgr.RefreshLock("acc-refresh-1")
			defer unlock()

			mu.Lock()
			activeCount++
			if activeCount > maxActive {
				maxActive = activeCount
			}
			mu.Unlock()

			time.Sleep(10 * time.Millisecond)

			mu.Lock()
			activeCount--
			mu.Unlock()
		}()
	}

	wg.Wait()

	if maxActive != 1 {
		t.Errorf("expected mutual exclusion (maxActive 1), got %d", maxActive)
	}
}
