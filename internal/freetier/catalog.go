package freetier

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CatalogStatus string

const (
	StatusVerified    CatalogStatus = "verified"
	StatusUnverified  CatalogStatus = "unverified"
	StatusChanged     CatalogStatus = "changed"
	StatusExpired     CatalogStatus = "expired"
	StatusUnavailable CatalogStatus = "unavailable"
)

type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

type CatalogEntry struct {
	ID                  string        `json:"id"`
	ProviderName        string        `json:"provider_name"`
	Category            string        `json:"category"`
	OfficialWebsite     string        `json:"official_website"`
	DocsURL             string        `json:"docs_url"`
	PricingURL          string        `json:"pricing_url"`
	FreeTierStatus      string        `json:"free_tier_status"`
	FreeQuota           string        `json:"free_quota"`
	ResetInterval       string        `json:"reset_interval"`
	SupportedRegions    string        `json:"supported_regions"`
	SignupSteps         string        `json:"signup_steps"`
	AuthMethod          string        `json:"auth_method"`
	RequiredScopes      string        `json:"required_scopes"`
	ExpirationBehavior  string        `json:"expiration_behavior"`
	Restrictions        string        `json:"restrictions"`
	PaymentRequired     bool          `json:"payment_required"`
	PersonalKeyRequired bool          `json:"personal_key_required"`
	SandboxSupport      bool          `json:"sandbox_support"`
	VerifiedAt          *time.Time    `json:"verified_at,omitempty"`
	Confidence          Confidence    `json:"confidence"`
	Status              CatalogStatus `json:"status"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

type Source struct {
	ID         string     `json:"id"`
	CatalogID  string     `json:"catalog_id"`
	SourceURL  string     `json:"source_url"`
	SourceType string     `json:"source_type"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	Notes      string     `json:"notes"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CatalogFilter struct {
	Category   string
	AuthMethod string
	FreeTier   *bool
	Region     string
	Sandbox    *bool
	QuotaType  string
	Status     string
}

type CatalogStore struct {
	db *sql.DB
}

func NewCatalogStore(db *sql.DB) *CatalogStore {
	return &CatalogStore{db: db}
}

func (cs *CatalogStore) CreateEntry(ctx context.Context, e *CatalogEntry) error {
	if e.ID == "" {
		e.ID = "ft_" + uuid.NewString()[:8]
	}
	now := time.Now()
	e.CreatedAt = now
	e.UpdatedAt = now

	_, err := cs.db.ExecContext(ctx, `
INSERT INTO free_tier_catalog (id, provider_name, category, official_website, docs_url, pricing_url,
    free_tier_status, free_quota, reset_interval, supported_regions, signup_steps, auth_method,
    required_scopes, expiration_behavior, restrictions, payment_required, personal_key_required,
    sandbox_support, verified_at, confidence, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.ProviderName, e.Category, e.OfficialWebsite, e.DocsURL, e.PricingURL,
		e.FreeTierStatus, e.FreeQuota, e.ResetInterval, e.SupportedRegions, e.SignupSteps,
		e.AuthMethod, e.RequiredScopes, e.ExpirationBehavior, e.Restrictions,
		boolToInt(e.PaymentRequired), boolToInt(e.PersonalKeyRequired), boolToInt(e.SandboxSupport),
		e.VerifiedAt, e.Confidence, e.Status, e.CreatedAt, e.UpdatedAt)
	return err
}

func (cs *CatalogStore) GetEntry(ctx context.Context, id string) (*CatalogEntry, error) {
	var e CatalogEntry
	var payReq, keyReq, sandbox int
	var verifiedAt sql.NullTime
	err := cs.db.QueryRowContext(ctx, `
SELECT id, provider_name, category, official_website, docs_url, pricing_url,
    free_tier_status, free_quota, reset_interval, supported_regions, signup_steps, auth_method,
    required_scopes, expiration_behavior, restrictions, payment_required, personal_key_required,
    sandbox_support, verified_at, confidence, status, created_at, updated_at
FROM free_tier_catalog WHERE id = ?`, id).Scan(
		&e.ID, &e.ProviderName, &e.Category, &e.OfficialWebsite, &e.DocsURL, &e.PricingURL,
		&e.FreeTierStatus, &e.FreeQuota, &e.ResetInterval, &e.SupportedRegions, &e.SignupSteps,
		&e.AuthMethod, &e.RequiredScopes, &e.ExpirationBehavior, &e.Restrictions,
		&payReq, &keyReq, &sandbox, &verifiedAt, &e.Confidence, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get catalog entry %s: %w", id, err)
	}
	e.PaymentRequired = payReq == 1
	e.PersonalKeyRequired = keyReq == 1
	e.SandboxSupport = sandbox == 1
	if verifiedAt.Valid {
		e.VerifiedAt = &verifiedAt.Time
	}
	return &e, nil
}

func (cs *CatalogStore) ListEntries(ctx context.Context, filter *CatalogFilter) ([]CatalogEntry, error) {
	query := `
SELECT id, provider_name, category, official_website, docs_url, pricing_url,
    free_tier_status, free_quota, reset_interval, supported_regions, signup_steps, auth_method,
    required_scopes, expiration_behavior, restrictions, payment_required, personal_key_required,
    sandbox_support, verified_at, confidence, status, created_at, updated_at
FROM free_tier_catalog WHERE 1=1`
	var args []any

	if filter != nil {
		if filter.Category != "" {
			query += " AND category = ?"
			args = append(args, filter.Category)
		}
		if filter.AuthMethod != "" {
			query += " AND auth_method = ?"
			args = append(args, filter.AuthMethod)
		}
		if filter.FreeTier != nil && *filter.FreeTier {
			query += " AND free_tier_status != 'none'"
		}
		if filter.Sandbox != nil && *filter.Sandbox {
			query += " AND sandbox_support = 1"
		}
		if filter.Status != "" {
			query += " AND status = ?"
			args = append(args, filter.Status)
		}
	}
	query += " ORDER BY provider_name ASC"

	rows, err := cs.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list catalog entries: %w", err)
	}
	defer rows.Close()

	var entries []CatalogEntry
	for rows.Next() {
		var e CatalogEntry
		var payReq, keyReq, sandbox int
		var verifiedAt sql.NullTime
		if err := rows.Scan(
			&e.ID, &e.ProviderName, &e.Category, &e.OfficialWebsite, &e.DocsURL, &e.PricingURL,
			&e.FreeTierStatus, &e.FreeQuota, &e.ResetInterval, &e.SupportedRegions, &e.SignupSteps,
			&e.AuthMethod, &e.RequiredScopes, &e.ExpirationBehavior, &e.Restrictions,
			&payReq, &keyReq, &sandbox, &verifiedAt, &e.Confidence, &e.Status, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			continue
		}
		e.PaymentRequired = payReq == 1
		e.PersonalKeyRequired = keyReq == 1
		e.SandboxSupport = sandbox == 1
		if verifiedAt.Valid {
			e.VerifiedAt = &verifiedAt.Time
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (cs *CatalogStore) GetCategories(ctx context.Context) ([]string, error) {
	rows, err := cs.db.QueryContext(ctx,
		"SELECT DISTINCT category FROM free_tier_catalog ORDER BY category ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err == nil {
			cats = append(cats, c)
		}
	}
	return cats, nil
}

func (cs *CatalogStore) GetVerified(ctx context.Context) ([]CatalogEntry, error) {
	return cs.ListEntries(ctx, &CatalogFilter{Status: string(StatusVerified)})
}

func (cs *CatalogStore) AddSource(ctx context.Context, src *Source) error {
	if src.ID == "" {
		src.ID = "src_" + uuid.NewString()[:8]
	}
	_, err := cs.db.ExecContext(ctx, `
INSERT INTO free_tier_sources (id, catalog_id, source_url, source_type, verified_at, notes)
VALUES (?, ?, ?, ?, ?, ?)`, src.ID, src.CatalogID, src.SourceURL, src.SourceType, src.VerifiedAt, src.Notes)
	return err
}

func (cs *CatalogStore) GetSources(ctx context.Context, catalogID string) ([]Source, error) {
	rows, err := cs.db.QueryContext(ctx, `
SELECT id, catalog_id, source_url, source_type, verified_at, notes, created_at
FROM free_tier_sources WHERE catalog_id = ? ORDER BY created_at DESC`, catalogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var s Source
		var verifiedAt sql.NullTime
		if err := rows.Scan(&s.ID, &s.CatalogID, &s.SourceURL, &s.SourceType, &verifiedAt, &s.Notes, &s.CreatedAt); err != nil {
			continue
		}
		if verifiedAt.Valid {
			s.VerifiedAt = &verifiedAt.Time
		}
		sources = append(sources, s)
	}
	return sources, nil
}

func (cs *CatalogStore) UpdateStatus(ctx context.Context, id string, status CatalogStatus) error {
	_, err := cs.db.ExecContext(ctx,
		"UPDATE free_tier_catalog SET status = ?, updated_at = ? WHERE id = ?",
		status, time.Now(), id)
	return err
}

func (cs *CatalogStore) DeleteEntry(ctx context.Context, id string) error {
	_, err := cs.db.ExecContext(ctx, "DELETE FROM free_tier_catalog WHERE id = ?", id)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
