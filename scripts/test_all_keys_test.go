package scripts

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/proxy"
	_ "modernc.org/sqlite"
)

func TestCheckAllAccountsWithProxyManager(t *testing.T) {
	if os.Getenv("EKAROUTER_ENABLE_LIVE_PROVIDER_TESTS") != "true" {
		t.Skip("skipping live test; set EKAROUTER_ENABLE_LIVE_PROVIDER_TESTS=true")
	}
	dbPath := filepath.Join("..", "data", "ekarouter.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	keyPath := filepath.Join("..", "data", ".secret_key")
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read .secret_key: %v", err)
	}
	crypto, _ := auth.NewCryptoService(string(keyBytes))

	rows, err := db.Query(`
SELECT a.id, a.provider_id, a.name, a.enabled, COALESCE(c.encrypted_access, ''), COALESCE(c.encrypted_secret, '')
FROM accounts a
LEFT JOIN credentials c ON c.account_id = a.id
ORDER BY a.priority DESC`)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	defer rows.Close()

	type AccInfo struct {
		id         string
		providerID string
		name       string
		enabled    int
		encAccess  string
		encSecret  string
	}
	var list []AccInfo
	for rows.Next() {
		var ai AccInfo
		if err := rows.Scan(&ai.id, &ai.providerID, &ai.name, &ai.enabled, &ai.encAccess, &ai.encSecret); err == nil {
			list = append(list, ai)
		}
	}

	proxyMgr := proxy.NewManager(false)

	type Result struct {
		id      string
		prov    string
		name    string
		healthy bool
		status  int
		lat     int64
		msg     string
	}

	results := make([]Result, len(list))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for i, acc := range list {
		wg.Add(1)
		go func(idx int, a AccInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			apiKey := ""
			if a.encAccess != "" && crypto != nil {
				apiKey, _ = crypto.Decrypt(a.encAccess)
			}

			var meta map[string]any
			if a.encSecret != "" && crypto != nil {
				if sec, err := crypto.Decrypt(a.encSecret); err == nil {
					_ = json.Unmarshal([]byte(sec), &meta)
				}
			}

			if apiKey == "" {
				results[idx] = Result{id: a.id, prov: a.providerID, name: a.name, healthy: false, msg: "empty credential"}
				return
			}

			var provKind, provBaseURL string
			_ = db.QueryRow("SELECT COALESCE(kind, 'openai'), COALESCE(base_url, '') FROM providers WHERE id = ?", a.providerID).Scan(&provKind, &provBaseURL)
			if provBaseURL == "" {
				if provKind == "gemini" {
					provBaseURL = "https://generativelanguage.googleapis.com"
				} else {
					provBaseURL = "https://api.openai.com/v1"
				}
			}

			proxyPoolID, _ := meta["proxyPoolId"].(string)
			var prof *proxy.Profile
			if proxyPoolID != "" {
				var scheme, host string
				var port int
				var user, encPass sql.NullString
				if err := db.QueryRow("SELECT scheme, host, port, username, encrypted_password FROM proxy_profiles WHERE id = ?", proxyPoolID).Scan(&scheme, &host, &port, &user, &encPass); err == nil {
					pass := ""
					if encPass.Valid && crypto != nil {
						pass, _ = crypto.Decrypt(encPass.String)
					}
					prof = &proxy.Profile{
						ID:       proxyPoolID,
						Scheme:   scheme,
						Host:     host,
						Port:     port,
						Username: user.String,
						Password: pass,
					}
				}
			}

			client, _ := proxyMgr.GetClient(prof, 8*time.Second)
			if client == nil {
				client = &http.Client{Timeout: 8 * time.Second}
			}

			cleanedBase := strings.TrimRight(provBaseURL, "/")
			var testReq *http.Request
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()

			switch {
			case a.providerID == "gemini" || provKind == "gemini":
				testURL := fmt.Sprintf("%s/v1beta/models?key=%s", cleanedBase, apiKey)
				if strings.Contains(cleanedBase, "/models") {
					testURL = fmt.Sprintf("%s?key=%s", cleanedBase, apiKey)
				}
				testReq, _ = http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
			case a.providerID == "anthropic" || provKind == "anthropic":
				testURL := cleanedBase + "/messages"
				body := []byte(`{"model":"claude-3-5-haiku-20241022","max_tokens":1,"messages":[{"role":"user","content":"ping"}]}`)
				testReq, _ = http.NewRequestWithContext(ctx, http.MethodPost, testURL, strings.NewReader(string(body)))
				if testReq != nil {
					testReq.Header.Set("Content-Type", "application/json")
					testReq.Header.Set("x-api-key", apiKey)
					testReq.Header.Set("anthropic-version", "2023-06-01")
				}
			case a.providerID == "github":
				testURL := "https://models.inference.ai.azure.com/models"
				testReq, _ = http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
				if testReq != nil {
					testReq.Header.Set("Authorization", "Bearer "+apiKey)
				}
			default:
				testURL := cleanedBase + "/models"
				testReq, _ = http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
				if testReq != nil {
					testReq.Header.Set("Authorization", "Bearer "+apiKey)
				}
			}

			if testReq == nil {
				results[idx] = Result{id: a.id, prov: a.providerID, name: a.name, healthy: false, msg: "nil request"}
				return
			}

			start := time.Now()
			resp, err := client.Do(testReq)
			lat := time.Since(start).Milliseconds()

			if (err != nil || (resp != nil && resp.StatusCode == 404)) && prof != nil {
				if resp != nil {
					resp.Body.Close()
				}
				directClient := &http.Client{Timeout: 8 * time.Second}
				startDirect := time.Now()
				directReq, errDirect := http.NewRequestWithContext(context.Background(), testReq.Method, testReq.URL.String(), nil)
				if errDirect == nil {
					for k, v := range testReq.Header {
						directReq.Header[k] = v
					}
					respDirect, errD := directClient.Do(directReq)
					if errD == nil {
						resp = respDirect
						lat = time.Since(startDirect).Milliseconds()
						err = nil
					}
				}
			}

			if err != nil {
				results[idx] = Result{id: a.id, prov: a.providerID, name: a.name, healthy: false, lat: lat, msg: err.Error()}
				return
			}
			defer resp.Body.Close()

			healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
			msg := "OK"
			if !healthy {
				msg = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
			}
			results[idx] = Result{
				id:      a.id,
				prov:    a.providerID,
				name:    a.name,
				healthy: healthy,
				status:  resp.StatusCode,
				lat:     lat,
				msg:     msg,
			}
		}(i, acc)
	}
	wg.Wait()

	healthyCount := 0
	for _, r := range results {
		if r.healthy {
			healthyCount++
		}
	}

	t.Logf("=== RESULTS SUMMARY ===")
	t.Logf("Total accounts: %d | Healthy: %d | Failed: %d", len(results), healthyCount, len(results)-healthyCount)

	groups := make(map[string][]Result)
	for _, r := range results {
		groups[r.prov] = append(groups[r.prov], r)
	}

	for prov, resList := range groups {
		hCount := 0
		for _, r := range resList {
			if r.healthy {
				hCount++
			}
		}
		t.Logf("Provider %-25s: %2d / %2d healthy", prov, hCount, len(resList))
		for _, r := range resList {
			mark := "FAIL"
			if r.healthy {
				mark = "PASS"
			}
			if r.healthy || !strings.Contains(r.msg, "empty credential") {
				t.Logf("  [%s] %-20s | %4dms | %s", mark, r.name, r.lat, r.msg)
			}
		}
	}
}
