package audit

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"
)

type Record struct {
	ID            int64     `json:"id"`
	ActorID       string    `json:"actor_id"`
	ActorType     string    `json:"actor_type"`
	Action        string    `json:"action"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    string    `json:"resource_id"`
	ProjectID     string    `json:"project_id,omitempty"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	RequestID     string    `json:"request_id"`
	Result        string    `json:"result"`
	ErrorCategory string    `json:"error_category,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type Logger struct {
	db     *sql.DB
	ch     chan *Record
	wg     sync.WaitGroup
	closed bool
	mu     sync.Mutex
}

func NewLogger(db *sql.DB, bufferSize int) *Logger {
	if bufferSize <= 0 {
		bufferSize = 1000
	}
	l := &Logger{
		db: db,
		ch: make(chan *Record, bufferSize),
	}
	l.wg.Add(1)
	go l.worker()
	return l
}

func (l *Logger) Log(r *Record) {
	if r == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	if r.Result == "" {
		r.Result = "success"
	}
	if r.ActorType == "" {
		r.ActorType = "user"
	}

	select {
	case l.ch <- r:
	default:
	}
}

func (l *Logger) worker() {
	defer l.wg.Done()
	query := `
INSERT INTO audit_logs (
    actor_id, actor_type, action, resource_type, resource_id,
    project_id, ip_address, user_agent, request_id, result, error_category, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for r := range l.ch {
		var pID any
		if r.ProjectID != "" {
			pID = r.ProjectID
		}
		_, _ = l.db.Exec(query,
			r.ActorID, r.ActorType, r.Action, r.ResourceType, r.ResourceID,
			pID, r.IPAddress, r.UserAgent, r.RequestID, r.Result, r.ErrorCategory, r.CreatedAt,
		)
	}
}

func (l *Logger) Close() {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return
	}
	l.closed = true
	close(l.ch)
	l.mu.Unlock()
	l.wg.Wait()
}

func (l *Logger) Query(ctx context.Context, action, actorID, resourceType string, limit int) ([]*Record, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var sb strings.Builder
	sb.WriteString(`
SELECT id, actor_id, actor_type, action, resource_type, resource_id,
       COALESCE(project_id, ''), ip_address, user_agent, request_id, result, error_category, created_at
FROM audit_logs WHERE 1=1`)

	var args []any
	if action != "" {
		sb.WriteString(" AND action = ?")
		args = append(args, action)
	}
	if actorID != "" {
		sb.WriteString(" AND actor_id = ?")
		args = append(args, actorID)
	}
	if resourceType != "" {
		sb.WriteString(" AND resource_type = ?")
		args = append(args, resourceType)
	}
	sb.WriteString(" ORDER BY created_at DESC LIMIT ?")
	args = append(args, limit)

	rows, err := l.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(
			&r.ID, &r.ActorID, &r.ActorType, &r.Action, &r.ResourceType, &r.ResourceID,
			&r.ProjectID, &r.IPAddress, &r.UserAgent, &r.RequestID, &r.Result, &r.ErrorCategory, &r.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &r)
	}
	return list, rows.Err()
}
