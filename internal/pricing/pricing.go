package pricing

import (
	"database/sql"
	"time"
)

type ModelPricing struct {
	ID               string  `json:"id"`
	ModelID          string  `json:"model_id"`
	Provider         string  `json:"provider"`
	InputCostPer1M   float64 `json:"input_cost_per_1m"`
	OutputCostPer1M  float64 `json:"output_cost_per_1m"`
	ImageCostPerUnit float64 `json:"image_cost_per_unit"`
	Currency         string  `json:"currency"`
	EffectiveDate    string  `json:"effective_date"`
	Source           string  `json:"source,omitempty"`
}

type CostEstimate struct {
	Model        string  `json:"model"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	InputCost    float64 `json:"input_cost"`
	OutputCost   float64 `json:"output_cost"`
	TotalCost    float64 `json:"total_cost"`
	Currency     string  `json:"currency"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) List(provider string) ([]ModelPricing, error) {
	q := `SELECT id, model_id, provider, input_cost_per_1m, output_cost_per_1m, image_cost_per_unit, currency, effective_date, COALESCE(source,'') FROM model_pricing`
	args := []any{}
	if provider != "" {
		q += ` WHERE provider = ?`
		args = append(args, provider)
	}
	q += ` ORDER BY provider, model_id`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ModelPricing
	for rows.Next() {
		var p ModelPricing
		if err := rows.Scan(&p.ID, &p.ModelID, &p.Provider, &p.InputCostPer1M, &p.OutputCostPer1M, &p.ImageCostPerUnit, &p.Currency, &p.EffectiveDate, &p.Source); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if result == nil {
		result = []ModelPricing{}
	}
	return result, nil
}

func (s *Service) Upsert(p ModelPricing) error {
	if p.ID == "" {
		p.ID = p.Provider + "_" + p.ModelID + "_" + time.Now().Format("20060102")
	}
	if p.Currency == "" {
		p.Currency = "USD"
	}
	if p.EffectiveDate == "" {
		p.EffectiveDate = time.Now().Format("2006-01-02")
	}
	_, err := s.db.Exec(`INSERT INTO model_pricing (id,model_id,provider,input_cost_per_1m,output_cost_per_1m,image_cost_per_unit,currency,effective_date,source,updated_at)
        VALUES (?,?,?,?,?,?,?,?,?,datetime('now'))
        ON CONFLICT(model_id,provider) DO UPDATE SET
        input_cost_per_1m=excluded.input_cost_per_1m,
        output_cost_per_1m=excluded.output_cost_per_1m,
        image_cost_per_unit=excluded.image_cost_per_unit,
        currency=excluded.currency,
        effective_date=excluded.effective_date,
        source=excluded.source,
        updated_at=datetime('now')`,
		p.ID, p.ModelID, p.Provider, p.InputCostPer1M, p.OutputCostPer1M, p.ImageCostPerUnit, p.Currency, p.EffectiveDate, p.Source)
	return err
}

func (s *Service) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM model_pricing WHERE id=?`, id)
	return err
}

func (s *Service) EstimateCost(modelID string, inputTokens, outputTokens int) (*CostEstimate, error) {
	var p ModelPricing
	err := s.db.QueryRow(`SELECT model_id, provider, input_cost_per_1m, output_cost_per_1m, currency FROM model_pricing WHERE model_id = ? LIMIT 1`, modelID).
		Scan(&p.ModelID, &p.Provider, &p.InputCostPer1M, &p.OutputCostPer1M, &p.Currency)
	if err == sql.ErrNoRows {
		return &CostEstimate{Model: modelID, InputTokens: inputTokens, OutputTokens: outputTokens, Currency: "USD"}, nil
	}
	if err != nil {
		return nil, err
	}
	inputCost := float64(inputTokens) / 1_000_000 * p.InputCostPer1M
	outputCost := float64(outputTokens) / 1_000_000 * p.OutputCostPer1M
	return &CostEstimate{
		Model:        modelID,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		InputCost:    inputCost,
		OutputCost:   outputCost,
		TotalCost:    inputCost + outputCost,
		Currency:     p.Currency,
	}, nil
}
