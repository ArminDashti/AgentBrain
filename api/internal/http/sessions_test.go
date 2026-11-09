package httpapi

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSessionIngest(t *testing.T) {
	_, r, _ := setupTestServer(t, false)
	token := mustLogin(t, r)

	sessionJSON := `{
		"uuid": "test-session-123",
		"summary": {
			"name": "Test Session",
			"subtitle": "Test subtitle",
			"status": "active",
			"unifiedMode": "true",
			"createdAt": 1234567890000,
			"lastUpdatedAt": 1234567900000,
			"inputTokens": 1000,
			"outputTokens": 500,
			"cacheReadTokens": 200,
			"cacheWriteTokens": 100,
			"modelConfig": {"model": "gpt-4"}
		},
		"header": {
			"workspaceId": "ws123",
			"value": {
				"workspaceIdentifier": {
					"uri": {
						"fsPath": "/test/path"
					}
				},
				"contextUsagePercent": 45.5,
				"totalLinesAdded": 50,
				"totalLinesRemoved": 20,
				"filesChangedCount": 3
			}
		},
		"conversation": [
			{
				"bubbleId": "bubble1",
				"bubble": {
					"type": 1,
					"text": "Hello",
					"startedAtMs": 1000,
					"completedAtMs": 1100,
					"inputTokens": 10,
					"outputTokens": 5
				}
			},
			{
				"bubbleId": "bubble2",
				"bubble": {
					"type": 2,
					"text": "Hi there!",
					"startedAtMs": 2000,
					"completedAtMs": 2200,
					"inputTokens": 20,
					"outputTokens": 15,
					"thinkingText": "Considering response..."
				},
				"grouping": {
					"hasThinking": true,
					"thinkingDurationMs": 50
				}
			}
		],
		"extractedAt": "2024-01-01T00:00:00Z"
	}`

	req := httptest.NewRequest("POST", "/api/v1/sessions/ingest", strings.NewReader(sessionJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("ingest: want 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["ok"] != true {
		t.Errorf("want ok=true")
	}
	if resp["uuid"].(string) != "test-session-123" {
		t.Errorf("uuid mismatch")
	}
	if resp["turn_count"].(float64) != 2 {
		t.Errorf("want 2 turns, got %v", resp["turn_count"])
	}

	req = httptest.NewRequest("GET", "/api/v1/sessions/test-session-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("get session: want 200, got %d", w.Code)
	}

	var session map[string]any
	json.Unmarshal(w.Body.Bytes(), &session)
	if session["name"].(string) != "Test Session" {
		t.Errorf("session name mismatch")
	}
	if session["turn_count"].(float64) != 2 {
		t.Errorf("turn_count mismatch")
	}
	if session["thinking_count"].(float64) != 1 {
		t.Errorf("thinking_count mismatch")
	}
}

func TestListSessions(t *testing.T) {
	_, r, _ := setupTestServer(t, false)
	token := mustLogin(t, r)

	for i := 1; i <= 3; i++ {
		idx := string(rune('0' + i))
		sessionJSON := `{
			"uuid": "test-session-` + idx + `",
			"summary": {
				"name": "Session ` + idx + `",
				"status": "active",
				"inputTokens": 1000,
				"outputTokens": 500
			},
			"header": {
				"workspaceId": "ws123",
				"value": {
					"workspaceIdentifier": {
						"uri": {
							"fsPath": "/test"
						}
					},
					"contextUsagePercent": 45.5
				}
			},
			"conversation": []
		}`
		req := httptest.NewRequest("POST", "/api/v1/sessions/ingest", strings.NewReader(sessionJSON))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("ingest %d: %d", i, w.Code)
		}
	}

	req := httptest.NewRequest("GET", "/api/v1/sessions?limit=2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("list: want 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	sessions := resp["sessions"].([]any)
	if len(sessions) != 2 {
		t.Errorf("want 2 sessions with limit=2, got %d", len(sessions))
	}

	req = httptest.NewRequest("GET", "/api/v1/sessions?q=Session+1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &resp)
	sessions = resp["sessions"].([]any)
	if len(sessions) != 1 {
		t.Errorf("search: want 1 match, got %d", len(sessions))
	}
}

func TestGetSessionTurns(t *testing.T) {
	_, r, _ := setupTestServer(t, false)
	token := mustLogin(t, r)

	sessionJSON := `{
		"uuid": "turn-test-session",
		"summary": {
			"name": "Turn Test",
			"status": "active"
		},
		"header": {
			"workspaceId": "ws123",
			"value": {
				"workspaceIdentifier": {
					"uri": {
						"fsPath": "/test"
					}
				}
			}
		},
		"conversation": [
			{
				"bubbleId": "b1",
				"bubble": {
					"type": 1,
					"text": "Hello"
				}
			},
			{
				"bubbleId": "b2",
				"bubble": {
					"type": 2,
					"text": "Hi!",
					"thinkingText": "Thinking..."
				},
				"grouping": {
					"hasThinking": true
				}
			}
		]
	}`

	req := httptest.NewRequest("POST", "/api/v1/sessions/ingest", strings.NewReader(sessionJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	req = httptest.NewRequest("GET", "/api/v1/sessions/turn-test-session/turns", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("get turns: want 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	turns := resp["turns"].([]any)
	if len(turns) != 2 {
		t.Errorf("want 2 turns, got %d", len(turns))
	}

	turn1 := turns[0].(map[string]any)
	if turn1["has_thinking"].(bool) != false {
		t.Errorf("turn 1 should not have thinking")
	}
	turn2 := turns[1].(map[string]any)
	if turn2["has_thinking"].(bool) != true {
		t.Errorf("turn 2 should have thinking")
	}
}

func TestGetSessionThinking(t *testing.T) {
	_, r, _ := setupTestServer(t, false)
	token := mustLogin(t, r)

	sessionJSON := `{
		"uuid": "thinking-test-session",
		"summary": {
			"name": "Thinking Test",
			"status": "active"
		},
		"header": {
			"workspaceId": "ws123"
		},
		"conversation": [
			{
				"bubbleId": "b1",
				"bubble": {
					"type": 2,
					"text": "Response 1",
					"thinkingText": "First thought"
				},
				"grouping": {
					"hasThinking": true,
					"thinkingDurationMs": 100
				}
			},
			{
				"bubbleId": "b2",
				"bubble": {
					"type": 2,
					"text": "Response 2",
					"thinkingText": "Second thought"
				},
				"grouping": {
					"hasThinking": true,
					"thinkingDurationMs": 200
				}
			}
		]
	}`

	req := httptest.NewRequest("POST", "/api/v1/sessions/ingest", strings.NewReader(sessionJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	req = httptest.NewRequest("GET", "/api/v1/sessions/thinking-test-session/thinking", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("get thinking: want 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	thinking := resp["thinking"].([]any)
	if len(thinking) != 2 {
		t.Errorf("want 2 thinking entries, got %d", len(thinking))
	}
}

func TestIngestInvalidSession(t *testing.T) {
	_, r, _ := setupTestServer(t, false)
	token := mustLogin(t, r)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "invalid json",
			body: `{bad json`,
		},
		{
			name: "missing uuid",
			body: `{"summary": {}, "conversation": []}`,
		},
		{
			name: "empty body",
			body: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/sessions/ingest", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code == 200 {
				t.Errorf("expected error for invalid input, got 200")
			}
		})
	}
}
