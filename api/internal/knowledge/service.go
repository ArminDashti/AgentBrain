package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/timefmt"
	"github.com/google/uuid"
)

type StoreTarget string

const TargetSQLite StoreTarget = "sqlite"

func ValidKind(k string) bool {
	switch k {
	case "general_knowledge", "expertise", "resolved_issues", "task_guide":
		return true
	}
	return false
}

func ValidTarget(t string) bool {
	if t == "" {
		return true
	}
	return StoreTarget(t) == TargetSQLite
}

type Item struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Tags      []string       `json:"tags"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

type Service struct {
	DB *sql.DB
}

type scannable interface{ Scan(dest ...any) error }

func scanItem(row scannable) (*Item, error) {
	var it Item
	var tagsRaw, metaRaw string
	var createdAt, updatedAt string
	if err := row.Scan(&it.ID, &it.Kind, &it.Title, &it.Body, &tagsRaw, &metaRaw, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	it.CreatedAt = timefmt.Normalize(createdAt)
	it.UpdatedAt = timefmt.Normalize(updatedAt)
	it.Tags = []string{}
	_ = json.Unmarshal([]byte(tagsRaw), &it.Tags)
	if it.Tags == nil {
		it.Tags = []string{}
	}
	it.Metadata = map[string]any{}
	_ = json.Unmarshal([]byte(metaRaw), &it.Metadata)
	return &it, nil
}

func marshalTags(tags []string) (string, error) {
	if tags == nil {
		tags = []string{}
	}
	b, err := json.Marshal(tags)
	return string(b), err
}

func (s *Service) Create(ctx context.Context, kind, title, body string, tags []string, metadata map[string]any, target StoreTarget) (*Item, error) {
	if !ValidKind(kind) {
		return nil, fmt.Errorf("invalid kind")
	}
	if !ValidTarget(string(target)) {
		return nil, fmt.Errorf("invalid store_target (sqlite only)")
	}
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title required")
	}
	if tags == nil {
		tags = []string{}
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	tagsJSON, err := marshalTags(tags)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	now := timefmt.NowUTC()
	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO knowledge (id, kind, title, body, tags, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, kind, title, body, tagsJSON, string(metaJSON), now, now)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *Service) Get(ctx context.Context, id string) (*Item, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, kind, title, body, tags, metadata, created_at, updated_at
		FROM knowledge WHERE id = ?`, id)
	return scanItem(row)
}

func (s *Service) List(ctx context.Context, kind, q string, limit int) ([]Item, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, kind, title, body, tags, metadata, created_at, updated_at FROM knowledge WHERE 1=1`
	args := []any{}
	if kind != "" {
		query += ` AND kind = ?`
		args = append(args, kind)
	}
	if q != "" {
		query += ` AND (lower(title) LIKE ? OR lower(body) LIKE ?)`
		like := "%" + strings.ToLower(q) + "%"
		args = append(args, like, like)
	}
	query += ` ORDER BY updated_at DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *it)
	}
	if out == nil {
		out = []Item{}
	}
	return out, rows.Err()
}

func (s *Service) Delete(ctx context.Context, id string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM knowledge WHERE id = ?`, id)
	return err
}

type SearchMode string

const ModeStructured SearchMode = "structured"

func (s *Service) Search(ctx context.Context, query, kind string, mode SearchMode, limit int) ([]Item, error) {
	if limit <= 0 {
		limit = 20
	}
	switch mode {
	case ModeStructured, "", "semantic", "hybrid":
		return s.List(ctx, kind, query, limit)
	default:
		return nil, fmt.Errorf("invalid mode")
	}
}

type Stats struct {
	TotalKnowledge         int            `json:"total_knowledge"`
	ByKind                 map[string]int `json:"by_kind"`
	SessionCount           int            `json:"session_count"`
	AvgContextUsagePercent *float64       `json:"avg_context_usage_percent"`
	RecentKnowledge        []Item         `json:"recent_knowledge"`
}

func (s *Service) Stats(ctx context.Context) (*Stats, error) {
	st := &Stats{ByKind: map[string]int{
		"general_knowledge": 0,
		"expertise":         0,
		"resolved_issues":   0,
		"task_guide":        0,
	}}
	rows, err := s.DB.QueryContext(ctx, `SELECT kind, COUNT(*) FROM knowledge GROUP BY kind`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var kind string
		var n int
		if err := rows.Scan(&kind, &n); err != nil {
			return nil, err
		}
		st.ByKind[kind] = n
		st.TotalKnowledge += n
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions`).Scan(&st.SessionCount); err != nil {
		return nil, err
	}
	var avg sql.NullFloat64
	if err := s.DB.QueryRowContext(ctx, `
		SELECT AVG(context_usage_percent) FROM sessions WHERE context_usage_percent IS NOT NULL
	`).Scan(&avg); err != nil {
		return nil, err
	}
	if avg.Valid {
		v := avg.Float64
		st.AvgContextUsagePercent = &v
	}
	recent, err := s.List(ctx, "", "", 8)
	if err != nil {
		return nil, err
	}
	st.RecentKnowledge = recent
	return st, nil
}
