package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/auth"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/config"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/db"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/embed"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/knowledge"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/qdrant"
	"github.com/gin-gonic/gin"
)

func setupTestServer(t *testing.T, withQdrant bool) (*Handler, *gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	tmpDB := fmt.Sprintf("/tmp/agent_brain_test_%d.db", time.Now().UnixNano())
	t.Cleanup(func() { os.Remove(tmpDB) })

	sqlDB, err := db.Open(tmpDB)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	if err := db.Migrate(sqlDB, "../../migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ctx := context.Background()
	hash, _ := auth.HashPassword(db.DefaultPassword)
	if err := db.SeedDefaultUser(ctx, sqlDB, hash); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	cfg := config.Config{
		JWTSecret:        "test-secret",
		QdrantURL:        os.Getenv("QDRANT_URL"),
		QdrantCollection: "test_collection",
		EmbeddingDim:     64,
	}
	if cfg.QdrantURL == "" {
		cfg.QdrantURL = "http://127.0.0.1:6335"
	}

	qd := qdrant.New(cfg.QdrantURL, cfg.QdrantCollection, cfg.EmbeddingDim)
	if withQdrant {
		if err := qd.EnsureCollection(ctx); err != nil {
			t.Fatalf("ensure qdrant collection: %v", err)
		}
	}

	emb := embed.New(cfg.EmbeddingDim, "", "", "")
	svc := &knowledge.Service{DB: sqlDB, Qdrant: qd, Embed: emb}
	h := New(sqlDB, cfg, svc, qd)

	r := gin.New()
	r.GET("/health", h.Health)
	api := r.Group("/api/v1")
	{
		api.POST("/auth/login", h.Login)
		authed := api.Group("")
		authed.Use(auth.Middleware(cfg.JWTSecret))
		{
			authed.GET("/stats", h.Stats)
			authed.POST("/knowledge", h.CreateKnowledge)
			authed.GET("/knowledge", h.ListKnowledge)
			authed.POST("/knowledge/search", h.SearchKnowledge)
			authed.GET("/knowledge/:id", h.GetKnowledge)
			authed.DELETE("/knowledge/:id", h.DeleteKnowledge)
			authed.POST("/sessions/ingest", h.IngestSession)
			authed.GET("/sessions", h.ListSessions)
			authed.GET("/sessions/:uuid", h.GetSession)
			authed.GET("/sessions/:uuid/thinking", h.GetSessionThinking)
			authed.GET("/sessions/:uuid/turns", h.GetSessionTurns)
			authed.GET("/prompts", h.ListPrompts)
			authed.POST("/prompts", h.CreatePrompt)
			authed.GET("/prompts/:id", h.GetPrompt)
			authed.PATCH("/prompts/:id", h.UpdatePrompt)
			authed.DELETE("/prompts/:id", h.DeletePrompt)
		}
	}

	return h, r, ""
}

func mustLogin(t *testing.T, r *gin.Engine) string {
	t.Helper()
	body := `{"username":"armin","password":"dopadopa123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Fatalf("no token in response")
	}
	return token
}

func TestHealth(t *testing.T) {
	_, r, _ := setupTestServer(t, true)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("want 200, got %d", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["ok"] != true {
		t.Errorf("want ok=true, got %v", resp)
	}
	if resp["sqlite"] != true {
		t.Errorf("want sqlite=true")
	}
	if resp["qdrant"] != true {
		t.Errorf("want qdrant=true")
	}
}

func TestHealthQdrantDown(t *testing.T) {
	_, r, _ := setupTestServer(t, false)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 503 {
		t.Errorf("want 503 when qdrant down, got %d", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["qdrant"] != false {
		t.Errorf("want qdrant=false")
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "valid credentials",
			body:       `{"username":"armin","password":"dopadopa123"}`,
			wantStatus: 200,
		},
		{
			name:       "invalid username",
			body:       `{"username":"baduser","password":"dopadopa123"}`,
			wantStatus: 401,
			wantError:  "invalid credentials",
		},
		{
			name:       "invalid password",
			body:       `{"username":"armin","password":"wrongpass"}`,
			wantStatus: 401,
			wantError:  "invalid credentials",
		},
		{
			name:       "missing username",
			body:       `{"password":"dopadopa123"}`,
			wantStatus: 401,
		},
		{
			name:       "invalid json",
			body:       `{bad json`,
			wantStatus: 400,
			wantError:  "invalid body",
		},
		{
			name:       "empty body",
			body:       ``,
			wantStatus: 400,
		},
	}

	_, r, _ := setupTestServer(t, false)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
			if tt.wantError != "" {
				var resp map[string]any
				json.Unmarshal(w.Body.Bytes(), &resp)
				errMsg, _ := resp["error"].(string)
				if !strings.Contains(errMsg, tt.wantError) {
					t.Errorf("want error containing %q, got %q", tt.wantError, errMsg)
				}
			}
		})
	}
}

func TestKnowledgeCRUD(t *testing.T) {
	_, r, _ := setupTestServer(t, true)
	token := mustLogin(t, r)

	kinds := []string{"general_knowledge", "expertise", "resolved_issues", "task_guide"}
	targets := []string{"sqlite", "qdrant", "both"}

	for _, kind := range kinds {
		for _, target := range targets {
			t.Run(fmt.Sprintf("%s_%s", kind, target), func(t *testing.T) {
				body := fmt.Sprintf(`{
					"kind": "%s",
					"title": "Test %s",
					"body": "This is a test body for %s",
					"tags": ["test", "automated"],
					"metadata": {"test": true},
					"store_target": "%s"
				}`, kind, kind, kind, target)

				req := httptest.NewRequest("POST", "/api/v1/knowledge", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if w.Code != 201 {
					t.Fatalf("create: want 201, got %d: %s", w.Code, w.Body.String())
				}

				var created map[string]any
				json.Unmarshal(w.Body.Bytes(), &created)
				id := created["id"].(string)

				inSQLite := created["in_sqlite"].(bool)
				inQdrant := created["in_qdrant"].(bool)
				if target == "sqlite" && (!inSQLite || inQdrant) {
					t.Errorf("want in_sqlite=true, in_qdrant=false for target=sqlite")
				}
				if target == "qdrant" && (inSQLite || !inQdrant) {
					t.Errorf("want in_sqlite=false, in_qdrant=true for target=qdrant")
				}
				if target == "both" && (!inSQLite || !inQdrant) {
					t.Errorf("want both=true for target=both")
				}

				req = httptest.NewRequest("GET", "/api/v1/knowledge/"+id, nil)
				req.Header.Set("Authorization", "Bearer "+token)
				w = httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if w.Code != 200 {
					t.Fatalf("get: want 200, got %d", w.Code)
				}

				var got map[string]any
				json.Unmarshal(w.Body.Bytes(), &got)
				if got["kind"].(string) != kind {
					t.Errorf("want kind=%s, got %v", kind, got["kind"])
				}

				req = httptest.NewRequest("DELETE", "/api/v1/knowledge/"+id, nil)
				req.Header.Set("Authorization", "Bearer "+token)
				w = httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if w.Code != 200 {
					t.Fatalf("delete: want 200, got %d", w.Code)
				}

				req = httptest.NewRequest("GET", "/api/v1/knowledge/"+id, nil)
				req.Header.Set("Authorization", "Bearer "+token)
				w = httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if w.Code != 404 {
					t.Errorf("get after delete: want 404, got %d", w.Code)
				}
			})
		}
	}
}

func TestKnowledgeInvalidInput(t *testing.T) {
	_, r, _ := setupTestServer(t, true)
	token := mustLogin(t, r)

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid kind",
			body:       `{"kind":"invalid","title":"Test","body":"Test","store_target":"sqlite"}`,
			wantStatus: 400,
			wantError:  "invalid kind",
		},
		{
			name:       "invalid store_target",
			body:       `{"kind":"general_knowledge","title":"Test","body":"Test","store_target":"invalid"}`,
			wantStatus: 400,
			wantError:  "invalid store_target",
		},
		{
			name:       "missing title",
			body:       `{"kind":"general_knowledge","body":"Test","store_target":"sqlite"}`,
			wantStatus: 400,
			wantError:  "title required",
		},
		{
			name:       "empty title",
			body:       `{"kind":"general_knowledge","title":"   ","body":"Test","store_target":"sqlite"}`,
			wantStatus: 400,
			wantError:  "title required",
		},
		{
			name:       "invalid json",
			body:       `{bad json`,
			wantStatus: 400,
			wantError:  "invalid body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/knowledge", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
			var resp map[string]any
			json.Unmarshal(w.Body.Bytes(), &resp)
			errMsg, _ := resp["error"].(string)
			if !strings.Contains(errMsg, tt.wantError) {
				t.Errorf("want error containing %q, got %q", tt.wantError, errMsg)
			}
		})
	}
}

func TestKnowledgeUnicode(t *testing.T) {
	_, r, _ := setupTestServer(t, true)
	token := mustLogin(t, r)

	body := `{
		"kind": "general_knowledge",
		"title": "تست فارسی",
		"body": "این یک متن فارسی است 🎉 with emoji and العربية",
		"tags": ["فارسی", "test", "مختلط"],
		"metadata": {"language": "mixed"},
		"store_target": "both"
	}`

	req := httptest.NewRequest("POST", "/api/v1/knowledge", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}

	var created map[string]any
	json.Unmarshal(w.Body.Bytes(), &created)
	if created["title"].(string) != "تست فارسی" {
		t.Errorf("unicode title not preserved: %v", created["title"])
	}
	bodyStr := created["body"].(string)
	if !strings.Contains(bodyStr, "فارسی") || !strings.Contains(bodyStr, "🎉") {
		t.Errorf("unicode body not preserved: %v", bodyStr)
	}
}

func TestKnowledgeLargePayload(t *testing.T) {
	_, r, _ := setupTestServer(t, true)
	token := mustLogin(t, r)

	largeBody := strings.Repeat("This is a large body. ", 5000)
	body := fmt.Sprintf(`{
		"kind": "general_knowledge",
		"title": "Large payload test",
		"body": %q,
		"tags": [],
		"store_target": "sqlite"
	}`, largeBody)

	req := httptest.NewRequest("POST", "/api/v1/knowledge", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}

	var created map[string]any
	json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	req = httptest.NewRequest("GET", "/api/v1/knowledge/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var got map[string]any
	json.Unmarshal(w.Body.Bytes(), &got)
	if got["body"].(string) != largeBody {
		t.Errorf("large body not preserved")
	}
}

func TestKnowledgeSearch(t *testing.T) {
	_, r, _ := setupTestServer(t, true)
	token := mustLogin(t, r)

	testData := []map[string]string{
		{"kind": "general_knowledge", "title": "Go programming", "body": "Go is a statically typed language"},
		{"kind": "expertise", "title": "TypeScript best practices", "body": "TypeScript adds types to JavaScript"},
		{"kind": "resolved_issues", "title": "Fixed login bug", "body": "The authentication token was expired"},
	}

	for _, data := range testData {
		body := fmt.Sprintf(`{
			"kind": "%s",
			"title": "%s",
			"body": "%s",
			"tags": [],
			"store_target": "both"
		}`, data["kind"], data["title"], data["body"])
		req := httptest.NewRequest("POST", "/api/v1/knowledge", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 201 {
			t.Fatalf("create failed: %d", w.Code)
		}
	}

	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name       string
		searchBody string
		wantCount  int
	}{
		{
			name:       "structured search",
			searchBody: `{"query":"programming","mode":"structured","limit":10}`,
			wantCount:  1,
		},
		{
			name:       "semantic search",
			searchBody: `{"query":"coding languages","mode":"semantic","limit":10}`,
			wantCount:  2,
		},
		{
			name:       "hybrid search",
			searchBody: `{"query":"TypeScript","mode":"hybrid","limit":10}`,
			wantCount:  1,
		},
		{
			name:       "filter by kind",
			searchBody: `{"query":"","kind":"resolved_issues","mode":"structured","limit":10}`,
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/knowledge/search", strings.NewReader(tt.searchBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != 200 {
				t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
			}
			var resp map[string]any
			json.Unmarshal(w.Body.Bytes(), &resp)
			items := resp["items"].([]any)
			if len(items) < tt.wantCount {
				t.Errorf("want at least %d results, got %d", tt.wantCount, len(items))
			}
		})
	}
}

func TestPromptsCRUD(t *testing.T) {
	_, r, _ := setupTestServer(t, false)
	token := mustLogin(t, r)

	createBody := `{
		"prompt": "Write a function",
		"technically_revised": "Write a Go function",
		"english_revised": "Please write a function",
		"harness": "go test"
	}`

	req := httptest.NewRequest("POST", "/api/v1/prompts", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("create: want 201, got %d: %s", w.Code, w.Body.String())
	}

	var created map[string]any
	json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	req = httptest.NewRequest("GET", "/api/v1/prompts/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("get: want 200, got %d", w.Code)
	}

	var got map[string]any
	json.Unmarshal(w.Body.Bytes(), &got)
	if got["prompt"].(string) != "Write a function" {
		t.Errorf("prompt mismatch")
	}

	updateBody := `{
		"prompt": "Write a better function",
		"technically_revised": "Write an optimized Go function",
		"english_revised": "Please write an optimized function",
		"harness": "go test -v"
	}`

	req = httptest.NewRequest("PATCH", "/api/v1/prompts/"+id, strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("update: want 200, got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/api/v1/prompts/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &got)
	if got["prompt"].(string) != "Write a better function" {
		t.Errorf("update not applied")
	}

	req = httptest.NewRequest("DELETE", "/api/v1/prompts/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("delete: want 200, got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/api/v1/prompts/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("get after delete: want 404, got %d", w.Code)
	}
}

func TestListKnowledge(t *testing.T) {
	_, r, _ := setupTestServer(t, true)
	token := mustLogin(t, r)

	for i := 0; i < 5; i++ {
		body := fmt.Sprintf(`{
			"kind": "general_knowledge",
			"title": "Item %d",
			"body": "Body %d",
			"tags": [],
			"store_target": "sqlite"
		}`, i, i)
		req := httptest.NewRequest("POST", "/api/v1/knowledge", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 201 {
			t.Fatalf("create %d: %d", i, w.Code)
		}
	}

	req := httptest.NewRequest("GET", "/api/v1/knowledge?limit=3", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("list: want 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	items := resp["items"].([]any)
	if len(items) != 3 {
		t.Errorf("want 3 items with limit=3, got %d", len(items))
	}
}

func TestStats(t *testing.T) {
	_, r, _ := setupTestServer(t, true)
	token := mustLogin(t, r)

	body := `{
		"kind": "expertise",
		"title": "Test expertise",
		"body": "Test",
		"tags": [],
		"store_target": "sqlite"
	}`
	req := httptest.NewRequest("POST", "/api/v1/knowledge", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	req = httptest.NewRequest("GET", "/api/v1/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("stats: want 200, got %d", w.Code)
	}

	var stats map[string]any
	json.Unmarshal(w.Body.Bytes(), &stats)
	if stats["total_knowledge"].(float64) < 1 {
		t.Errorf("expected at least 1 knowledge item")
	}
	byKind := stats["by_kind"].(map[string]any)
	if byKind["expertise"].(float64) < 1 {
		t.Errorf("expected at least 1 expertise item")
	}
	if _, ok := stats["avg_context_usage_percent"]; !ok {
		t.Errorf("expected avg_context_usage_percent in stats response")
	}
}

func TestUnauthorized(t *testing.T) {
	_, r, _ := setupTestServer(t, false)

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/knowledge"},
		{"POST", "/api/v1/knowledge"},
		{"GET", "/api/v1/stats"},
		{"GET", "/api/v1/prompts"},
		{"POST", "/api/v1/prompts"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+"_"+ep.path, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != 401 {
				t.Errorf("want 401 without auth, got %d", w.Code)
			}
		})
	}
}

func TestConcurrentWrites(t *testing.T) {
	_, r, _ := setupTestServer(t, true)
	token := mustLogin(t, r)

	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			body := fmt.Sprintf(`{
				"kind": "general_knowledge",
				"title": "Concurrent %d",
				"body": "Test %d",
				"tags": [],
				"store_target": "sqlite"
			}`, idx, idx)
			req := httptest.NewRequest("POST", "/api/v1/knowledge", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != 201 {
				done <- fmt.Errorf("goroutine %d: want 201, got %d", idx, w.Code)
				return
			}
			done <- nil
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		if err := <-done; err != nil {
			t.Error(err)
		}
	}
}
