package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/platform"
	"github.com/google/uuid"
)

type Webhook struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	TargetURL string    `json:"target_url"`
	Events    string    `json:"events"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Delivery struct {
	ID             string    `json:"id"`
	WebhookID      string    `json:"webhook_id"`
	Event          string    `json:"event"`
	PayloadJSON    string    `json:"payload_json"`
	ResponseStatus int       `json:"response_status"`
	LatencyMS      int64     `json:"latency_ms"`
	Error          string    `json:"error,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Manager struct {
	db         *sql.DB
	httpClient *http.Client
	allowLocal bool
}

func NewManager(db *sql.DB, allowLocal bool) *Manager {
	return &Manager{
		db: db,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		allowLocal: allowLocal,
	}
}

func GenerateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return "whsec_" + hex.EncodeToString(b), nil
}

func SignPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifySignature(secret string, payload []byte, signature string) bool {
	expectedSig := SignPayload(secret, payload)
	sigToCompare := strings.TrimPrefix(signature, "sha256=")
	return hmac.Equal([]byte(sigToCompare), []byte(expectedSig))
}

func (m *Manager) CreateWebhook(ctx context.Context, name, targetURL, events, encryptedSecret string) (*Webhook, error) {
	if name == "" || targetURL == "" || encryptedSecret == "" {
		return nil, errors.New("name, target_url, and secret are required")
	}
	if !m.allowLocal {
		if err := platform.ValidateSSRF(targetURL); err != nil {
			return nil, fmt.Errorf("SSRF rejection: %w", err)
		}
	}

	wh := &Webhook{
		ID:        uuid.New().String(),
		Name:      name,
		TargetURL: targetURL,
		Events:    events,
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	query := `
INSERT INTO webhooks (id, name, target_url, secret_encrypted, events, enabled, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 1, ?, ?)`

	_, err := m.db.ExecContext(ctx, query, wh.ID, wh.Name, wh.TargetURL, encryptedSecret, wh.Events, wh.CreatedAt, wh.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return wh, nil
}

func (m *Manager) ListWebhooks(ctx context.Context) ([]*Webhook, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT id, name, target_url, events, enabled, created_at, updated_at FROM webhooks ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Webhook
	for rows.Next() {
		var wh Webhook
		var enInt int
		if err := rows.Scan(&wh.ID, &wh.Name, &wh.TargetURL, &wh.Events, &enInt, &wh.CreatedAt, &wh.UpdatedAt); err != nil {
			return nil, err
		}
		wh.Enabled = enInt == 1
		list = append(list, &wh)
	}
	return list, rows.Err()
}

func (m *Manager) Dispatch(ctx context.Context, webhookID, rawSecret string, event string, payload []byte) (*Delivery, error) {
	var targetURL string
	var enabledInt int
	err := m.db.QueryRowContext(ctx, "SELECT target_url, enabled FROM webhooks WHERE id = ?", webhookID).Scan(&targetURL, &enabledInt)
	if err != nil {
		return nil, fmt.Errorf("webhook not found: %w", err)
	}
	if enabledInt == 0 {
		return nil, errors.New("webhook is disabled")
	}

	if !m.allowLocal {
		if err := platform.ValidateSSRF(targetURL); err != nil {
			return nil, fmt.Errorf("SSRF rejection: %w", err)
		}
	}

	signature := SignPayload(rawSecret, payload)
	deliveryID := uuid.New().String()
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "EkaRouter-Webhook/1.0")
	req.Header.Set("X-EkaRouter-Event", event)
	req.Header.Set("X-EkaRouter-Delivery", deliveryID)
	req.Header.Set("X-EkaRouter-Signature", "sha256="+signature)

	var respStatus int
	var respErr string

	resp, err := m.httpClient.Do(req)
	latency := time.Since(start)

	if err != nil {
		respErr = err.Error()
	} else {
		respStatus = resp.StatusCode
		_ = resp.Body.Close()
		if resp.StatusCode >= 400 {
			respErr = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
	}

	del := &Delivery{
		ID:             deliveryID,
		WebhookID:      webhookID,
		Event:          event,
		PayloadJSON:    string(payload),
		ResponseStatus: respStatus,
		LatencyMS:      latency.Milliseconds(),
		Error:          respErr,
		CreatedAt:      time.Now().UTC(),
	}

	_, _ = m.db.ExecContext(ctx, `
INSERT INTO webhook_deliveries (id, webhook_id, event, payload_json, response_status, latency_ms, error, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, del.ID, del.WebhookID, del.Event, del.PayloadJSON, del.ResponseStatus, del.LatencyMS, del.Error, del.CreatedAt)

	return del, err
}

func (m *Manager) ListDeliveries(ctx context.Context, webhookID string) ([]*Delivery, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT id, webhook_id, event, payload_json, response_status, latency_ms, error, created_at FROM webhook_deliveries WHERE webhook_id = ? ORDER BY created_at DESC LIMIT 50", webhookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Delivery
	for rows.Next() {
		var d Delivery
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.Event, &d.PayloadJSON, &d.ResponseStatus, &d.LatencyMS, &d.Error, &d.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}
