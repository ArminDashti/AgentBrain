package httpapi

import (
	"database/sql"
	"net/http"

	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/timefmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type promptRow struct {
	ID                 string `json:"id"`
	Prompt             string `json:"prompt"`
	TechnicallyRevised string `json:"technically_revised"`
	EnglishRevised     string `json:"english_revised"`
	Harness            string `json:"harness"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type promptWriteReq struct {
	Prompt             string `json:"prompt"`
	TechnicallyRevised string `json:"technically_revised"`
	EnglishRevised     string `json:"english_revised"`
	Harness            string `json:"harness"`
}

func scanPrompt(row interface{ Scan(dest ...any) error }) (promptRow, error) {
	var p promptRow
	err := row.Scan(
		&p.ID, &p.Prompt, &p.TechnicallyRevised, &p.EnglishRevised, &p.Harness,
		&p.CreatedAt, &p.UpdatedAt,
	)
	p.CreatedAt = timefmt.Normalize(p.CreatedAt)
	p.UpdatedAt = timefmt.Normalize(p.UpdatedAt)
	return p, err
}

func (h *Handler) ListPrompts(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT id, prompt, technically_revised, english_revised, harness, created_at, updated_at
		FROM prompts
		ORDER BY updated_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []promptRow{}
	for rows.Next() {
		p, err := scanPrompt(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, p)
	}
	c.JSON(http.StatusOK, gin.H{"prompts": out})
}

func (h *Handler) GetPrompt(c *gin.Context) {
	id := c.Param("id")
	row := h.db.QueryRowContext(c.Request.Context(), `
		SELECT id, prompt, technically_revised, english_revised, harness, created_at, updated_at
		FROM prompts WHERE id = ?
	`, id)
	p, err := scanPrompt(row)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) CreatePrompt(c *gin.Context) {
	var req promptWriteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	id := uuid.NewString()
	now := timefmt.NowUTC()
	_, err := h.db.ExecContext(c.Request.Context(), `
		INSERT INTO prompts (id, prompt, technically_revised, english_revised, harness, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, req.Prompt, req.TechnicallyRevised, req.EnglishRevised, req.Harness, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *Handler) UpdatePrompt(c *gin.Context) {
	id := c.Param("id")
	var req promptWriteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	now := timefmt.NowUTC()
	res, err := h.db.ExecContext(c.Request.Context(), `
		UPDATE prompts
		SET prompt = ?, technically_revised = ?, english_revised = ?, harness = ?, updated_at = ?
		WHERE id = ?
	`, req.Prompt, req.TechnicallyRevised, req.EnglishRevised, req.Harness, now, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) DeletePrompt(c *gin.Context) {
	res, err := h.db.ExecContext(c.Request.Context(), `DELETE FROM prompts WHERE id = ?`, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
