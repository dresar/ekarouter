package quota

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/db"
)

type ModelQuota struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Used                int     `json:"used"`
	Total               int     `json:"total"`
	RemainingPercentage float64 `json:"remaining_percentage"`
	ResetAt             string  `json:"reset_at,omitempty"`
	DisplayName         string  `json:"display_name,omitempty"`
}

type AccountQuota struct {
	AccountID        string       `json:"account_id"`
	AccountName      string       `json:"account_name"`
	ProviderID       string       `json:"provider_id"`
	ProviderName     string       `json:"provider_name"`
	AuthType         string       `json:"auth_type"`
	Email            string       `json:"email,omitempty"`
	State            string       `json:"state"`
	Plan             string       `json:"plan,omitempty"`
	OverallRemaining float64      `json:"overall_remaining"`
	ResetAt          string       `json:"reset_at,omitempty"`
	Quotas           []ModelQuota `json:"quotas"`
	LastChecked      time.Time    `json:"last_checked"`
}

type cachedItem struct {
	quota     *AccountQuota
	expiresAt time.Time
}

type Tracker struct {
	db     *db.DB
	crypto *auth.CryptoService
	client *http.Client
	cache  sync.Map
}

func NewTracker(database *db.DB, crypto *auth.CryptoService, client *http.Client) *Tracker {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &Tracker{
		db:     database,
		crypto: crypto,
		client: client,
	}
}

func (t *Tracker) GetAllQuotas(ctx context.Context, force bool) ([]AccountQuota, error) {
	rows, err := t.db.QueryContext(ctx, `
		SELECT a.id, a.provider_id, a.name, a.auth_type, a.state,
		       c.encrypted_access, c.encrypted_refresh, p.name as provider_name
		FROM accounts a
		JOIN providers p ON p.id = a.provider_id
		LEFT JOIN credentials c ON c.account_id = a.id
		WHERE a.enabled = 1
		ORDER BY a.priority ASC, a.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query accounts for quota: %w", err)
	}
	defer rows.Close()

	type accRow struct {
		id           string
		providerID   string
		name         string
		authType     string
		state        string
		encAccess    sql.NullString
		encRefresh   sql.NullString
		providerName string
	}

	var accs []accRow
	for rows.Next() {
		var r accRow
		if err := rows.Scan(&r.id, &r.providerID, &r.name, &r.authType, &r.state, &r.encAccess, &r.encRefresh, &r.providerName); err == nil {
			accs = append(accs, r)
		}
	}

	var wg sync.WaitGroup
	results := make([]AccountQuota, len(accs))
	sem := make(chan struct{}, 8)

	for i, acc := range accs {
		wg.Add(1)
		go func(idx int, a accRow) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if !force {
				if val, ok := t.cache.Load(a.id); ok {
					item := val.(cachedItem)
					if time.Now().Before(item.expiresAt) {
						results[idx] = *item.quota
						return
					}
				}
			}

			q := t.fetchSingleQuota(ctx, a.id, a.providerID, a.name, a.providerName, a.authType, a.state, a.encAccess.String)
			t.cache.Store(a.id, cachedItem{
				quota:     &q,
				expiresAt: time.Now().Add(45 * time.Second),
			})
			results[idx] = q
		}(i, acc)
	}

	wg.Wait()

	var final []AccountQuota
	for _, r := range results {
		if r.AccountID != "" {
			final = append(final, r)
		}
	}

	return final, nil
}

func (t *Tracker) GetAccountQuota(ctx context.Context, accountID string, force bool) (*AccountQuota, error) {
	if !force {
		if val, ok := t.cache.Load(accountID); ok {
			item := val.(cachedItem)
			if time.Now().Before(item.expiresAt) {
				return item.quota, nil
			}
		}
	}

	var providerID, name, authType, state, encAccess, providerName string
	err := t.db.QueryRowContext(ctx, `
		SELECT a.provider_id, a.name, a.auth_type, a.state,
		       COALESCE(c.encrypted_access, ''), p.name
		FROM accounts a
		JOIN providers p ON p.id = a.provider_id
		LEFT JOIN credentials c ON c.account_id = a.id
		WHERE a.id = ?
	`, accountID).Scan(&providerID, &name, &authType, &state, &encAccess, &providerName)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	q := t.fetchSingleQuota(ctx, accountID, providerID, name, providerName, authType, state, encAccess)
	t.cache.Store(accountID, cachedItem{
		quota:     &q,
		expiresAt: time.Now().Add(45 * time.Second),
	})

	return &q, nil
}

func (t *Tracker) fetchSingleQuota(ctx context.Context, accID, providerID, accName, providerName, authType, state, encAccess string) AccountQuota {
	token := ""
	if encAccess != "" && t.crypto != nil {
		if decrypted, err := t.crypto.Decrypt(encAccess); err == nil {
			token = decrypted
		}
	}

	base := AccountQuota{
		AccountID:        accID,
		AccountName:      accName,
		ProviderID:       providerID,
		ProviderName:     providerName,
		AuthType:         authType,
		State:            state,
		Plan:             "Standard",
		OverallRemaining: 100,
		LastChecked:      time.Now(),
		Quotas:           make([]ModelQuota, 0),
	}

	normPID := strings.ToLower(providerID)
	if (normPID == "antigravity" || normPID == "gemini-agy") && token != "" {
		return t.fetchAntigravityQuota(ctx, base, token)
	}

	if normPID == "gemini-cli" && token != "" {
		return t.fetchGeminiCLIQuota(ctx, base, token)
	}

	return t.fetchGenericUsageQuota(ctx, base)
}

func (t *Tracker) fetchAntigravityQuota(ctx context.Context, base AccountQuota, token string) AccountQuota {
	base.Plan = "Pro Tier"

	reqBody := []byte(`{}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://daily-cloudcode-pa.googleapis.com/v1internal:fetchAvailableModels", bytes.NewReader(reqBody))
	if err != nil {
		return base
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "antigravity/ide/2.11.0 darwin/arm64")
	req.Header.Set("X-Client-Name", "antigravity")
	req.Header.Set("X-Client-Version", "2.11.0")

	resp, err := t.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return base
	}
	defer resp.Body.Close()

	var data struct {
		Models map[string]struct {
			DisplayName string `json:"displayName"`
			QuotaInfo   *struct {
				RemainingFraction float64 `json:"remainingFraction"`
				ResetTime         string  `json:"resetTime"`
			} `json:"quotaInfo"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return base
	}

	important := map[string]string{
		"gemini-3.8-flash-high":   "Gemini 3.8 Flash (High)",
		"gemini-3.8-flash-medium": "Gemini 3.8 Flash (Medium)",
		"gemini-3.8-flash-low":    "Gemini 3.8 Flash (Low)",
		"gemini-3.7-flash-high":   "Gemini 3.7 Flash (High)",
		"gemini-3.7-flash-medium": "Gemini 3.7 Flash (Medium)",
		"gemini-3.6-flash-high":   "Gemini 3.6 Flash (High)",
		"claude-sonnet-4-6":       "Claude Sonnet 4.6 (Thinking)",
		"claude-opus-4-6-thinking": "Claude Opus 4.6 (Thinking)",
		"gpt-oss-120b-medium":     "GPT-OSS 120B (Medium)",
	}

	var sumRemaining float64
	var count int
	var earliestReset string

	for k, label := range important {
		if m, ok := data.Models[k]; ok && m.QuotaInfo != nil {
			remFrac := m.QuotaInfo.RemainingFraction
			remPct := remFrac * 100.0
			total := 1000
			remVal := int(float64(total) * remFrac)
			usedVal := total - remVal

			dName := m.DisplayName
			if dName == "" {
				dName = label
			}

			base.Quotas = append(base.Quotas, ModelQuota{
				ID:                  k,
				Name:                label,
				DisplayName:         dName,
				Used:                usedVal,
				Total:               total,
				RemainingPercentage: remPct,
				ResetAt:             m.QuotaInfo.ResetTime,
			})

			sumRemaining += remPct
			count++

			if earliestReset == "" || (m.QuotaInfo.ResetTime != "" && m.QuotaInfo.ResetTime < earliestReset) {
				earliestReset = m.QuotaInfo.ResetTime
			}
		}
	}

	t.fetchAntigravityWeekly(ctx, &base, token)

	if count > 0 {
		base.OverallRemaining = sumRemaining / float64(count)
	}
	base.ResetAt = earliestReset

	return base
}

func (t *Tracker) fetchAntigravityWeekly(ctx context.Context, base *AccountQuota, token string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://daily-cloudcode-pa.googleapis.com/v1internal:retrieveUserQuotaSummary", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "antigravity/ide/2.11.0 darwin/arm64")

	resp, err := t.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return
	}
	defer resp.Body.Close()

	var data struct {
		Groups []struct {
			DisplayName string `json:"displayName"`
			Buckets     []struct {
				RemainingFraction float64 `json:"remainingFraction"`
				ResetTime         string  `json:"resetTime"`
			} `json:"buckets"`
		} `json:"groups"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
		for _, g := range data.Groups {
			if len(g.Buckets) > 0 {
				b := g.Buckets[0]
				pct := b.RemainingFraction * 100.0
				total := 1000
				rem := int(float64(total) * b.RemainingFraction)
				base.Quotas = append(base.Quotas, ModelQuota{
					ID:                  strings.ToLower(strings.ReplaceAll(g.DisplayName, " ", "_")),
					Name:                g.DisplayName,
					DisplayName:         g.DisplayName,
					Used:                total - rem,
					Total:               total,
					RemainingPercentage: pct,
					ResetAt:             b.ResetTime,
				})
			}
		}
	}
}

func (t *Tracker) fetchGeminiCLIQuota(ctx context.Context, base AccountQuota, token string) AccountQuota {
	base.Plan = "Free Tier"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://cloudaicompanion.googleapis.com/v1:retrieveUserQuota", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		return base
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return base
	}
	defer resp.Body.Close()

	var data struct {
		Buckets []struct {
			ModelID           string  `json:"modelId"`
			RemainingFraction float64 `json:"remainingFraction"`
			ResetTime         string  `json:"resetTime"`
		} `json:"buckets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
		var sum float64
		var count int
		for _, b := range data.Buckets {
			pct := b.RemainingFraction * 100.0
			total := 1000
			rem := int(float64(total) * b.RemainingFraction)
			base.Quotas = append(base.Quotas, ModelQuota{
				ID:                  b.ModelID,
				Name:                b.ModelID,
				DisplayName:         b.ModelID,
				Used:                total - rem,
				Total:               total,
				RemainingPercentage: pct,
				ResetAt:             b.ResetTime,
			})
			sum += pct
			count++
		}
		if count > 0 {
			base.OverallRemaining = sum / float64(count)
		}
	}

	return base
}

func (t *Tracker) fetchGenericUsageQuota(ctx context.Context, base AccountQuota) AccountQuota {
	today := time.Now().Format("2006-01-02")
	var reqCount, tokenCount int
	row := t.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(value), 0) FROM usage_records
		WHERE reference_type = 'account' AND reference_id = ? AND metric = 'requests' AND period_date = ?
	`, base.AccountID, today)
	_ = row.Scan(&reqCount)

	rowTokens := t.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(value), 0) FROM usage_records
		WHERE reference_type = 'account' AND reference_id = ? AND metric = 'tokens' AND period_date = ?
	`, base.AccountID, today)
	_ = rowTokens.Scan(&tokenCount)

	totalReqLimit := 10000
	remReq := totalReqLimit - reqCount
	if remReq < 0 {
		remReq = 0
	}
	remPct := (float64(remReq) / float64(totalReqLimit)) * 100.0

	base.OverallRemaining = remPct
	base.Quotas = append(base.Quotas, ModelQuota{
		ID:                  "daily_requests",
		Name:                "Daily Requests",
		DisplayName:         "Daily Requests",
		Used:                reqCount,
		Total:               totalReqLimit,
		RemainingPercentage: remPct,
		ResetAt:             time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour).Format(time.RFC3339),
	})

	return base
}
