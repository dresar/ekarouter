package combos

import (
	"database/sql"
)

type Combo struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Strategy    string       `json:"strategy"`
	StickyLimit int          `json:"sticky_limit"`
	Enabled     bool         `json:"enabled"`
	Models      []ComboModel `json:"models,omitempty"`
	CreatedAt   string       `json:"created_at"`
}

type ComboModel struct {
	ID       string `json:"id"`
	ComboID  string `json:"combo_id"`
	Model    string `json:"model"`
	Priority int    `json:"priority"`
	Weight   int    `json:"weight"`
	Enabled  bool   `json:"enabled"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) List() ([]Combo, error) {
	rows, err := s.db.Query(`SELECT id, name, COALESCE(description,''), strategy, sticky_limit, enabled, created_at FROM combos ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Combo
	for rows.Next() {
		var c Combo
		var enabled int
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Strategy, &c.StickyLimit, &enabled, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Enabled = enabled == 1
		result = append(result, c)
	}
	if result == nil {
		result = []Combo{}
	}
	return result, nil
}

func (s *Service) Get(id string) (*Combo, error) {
	var c Combo
	var enabled int
	err := s.db.QueryRow(`SELECT id, name, COALESCE(description,''), strategy, sticky_limit, enabled, created_at FROM combos WHERE id=?`, id).
		Scan(&c.ID, &c.Name, &c.Description, &c.Strategy, &c.StickyLimit, &enabled, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	c.Enabled = enabled == 1
	models, err := s.ListModels(c.ID)
	if err != nil {
		return nil, err
	}
	c.Models = models
	return &c, nil
}

func (s *Service) ListModels(comboID string) ([]ComboModel, error) {
	rows, err := s.db.Query(`SELECT id, combo_id, model, priority, weight, enabled FROM combo_models WHERE combo_id=? ORDER BY priority DESC`, comboID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ComboModel
	for rows.Next() {
		var m ComboModel
		var enabled int
		if err := rows.Scan(&m.ID, &m.ComboID, &m.Model, &m.Priority, &m.Weight, &enabled); err != nil {
			return nil, err
		}
		m.Enabled = enabled == 1
		result = append(result, m)
	}
	if result == nil {
		result = []ComboModel{}
	}
	return result, nil
}

func (s *Service) GetModelNames(comboName string) ([]string, error) {
	rows, err := s.db.Query(`SELECT cm.model FROM combo_models cm JOIN combos c ON c.id = cm.combo_id WHERE c.name = ? AND cm.enabled = 1 ORDER BY cm.priority DESC`, comboName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var models []string
	for rows.Next() {
		var m string
		if err := rows.Scan(&m); err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, nil
}

func (s *Service) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM combos WHERE id=?`, id)
	return err
}
