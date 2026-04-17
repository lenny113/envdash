package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	model "assignment-2/internal/models"
)

func newHandler(store *MockStore) *Handler {
	return newMockHandler(store)
}

// Helpers

type errorResponse struct {
	Error string `json:"error"`
}

// newHandler constructs a Handler wired to the provided mock store.
// Adjust this to match whatever constructor your package exposes.

func postAuthRequest(body interface{}) *http.Request {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/envdash/v1/auth/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func deleteAuthRequest(key string, apiKeyHeader string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, "/envdash/v1/auth/"+key, nil)
	req.SetPathValue("id", key)
	if apiKeyHeader != "" {
		req.Header.Set("X-Api-Key", apiKeyHeader)
	}
	return req
}

// decodeBody is a small helper to decode JSON response bodies in tests.
func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, dst interface{}) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(dst); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

// POST /auth/ — RegisterAuth

func TestRegisterAuth(t *testing.T) {
	type requestBody struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	tests := []struct {
		name           string
		method         string
		body           interface{}
		store          *MockStore
		wantStatusCode int
		// optional: inspect part of the response body
		wantKey bool // true → response must contain a non-empty "key" field
	}{
		{
			name:   "happy path — valid name and email",
			method: http.MethodPost,
			body:   requestBody{Name: "Alice", Email: "alice@example.com"},
			store: &MockStore{
				CountApiPerUserFn:  func(_ context.Context, _ string) (int, error) { return 0, nil },
				ApiKeyExistsFn:     func(ctx context.Context, key string) bool { return key == "valid" },
				CreateApiStorageFn: func(_ context.Context, _ model.Authentication) error { return nil },
			},
			wantStatusCode: http.StatusCreated,
			wantKey:        true,
		},
		{
			name:           "wrong method — GET not allowed",
			method:         http.MethodGet,
			body:           requestBody{Name: "Alice", Email: "alice@example.com"},
			store:          &MockStore{},
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "missing email",
			method:         http.MethodPost,
			body:           requestBody{Name: "Bob", Email: ""},
			store:          &MockStore{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "missing name",
			method:         http.MethodPost,
			body:           requestBody{Name: "", Email: "bob@example.com"},
			store:          &MockStore{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid email format — no @",
			method:         http.MethodPost,
			body:           requestBody{Name: "Charlie", Email: "notanemail"},
			store:          &MockStore{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid email format — trailing dot",
			method:         http.MethodPost,
			body:           requestBody{Name: "Charlie", Email: "charlie@.com"},
			store:          &MockStore{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON body",
			method:         http.MethodPost,
			body:           "this is not json",
			store:          &MockStore{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "firestore unreachable when counting keys",
			method: http.MethodPost,
			body:   requestBody{Name: "Dave", Email: "dave@example.com"},
			store: &MockStore{
				CountApiPerUserFn: func(_ context.Context, _ string) (int, error) {
					return 0, errors.New("firestore unavailable")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "user already has max API keys",
			method: http.MethodPost,
			body:   requestBody{Name: "Eve", Email: "eve@example.com"},
			store: &MockStore{
				// Return MAXAPIKEYS (assumed 5 here; adjust to match utils.MAXAPIKEYS)
				CountApiPerUserFn: func(_ context.Context, _ string) (int, error) { return 5, nil },
			},
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name:   "firestore write fails",
			method: http.MethodPost,
			body:   requestBody{Name: "Frank", Email: "frank@example.com"},
			store: &MockStore{
				CountApiPerUserFn: func(_ context.Context, _ string) (int, error) { return 0, nil },
				ApiKeyExistsFn:    func(_ context.Context, _ string) bool { return false },
				CreateApiStorageFn: func(_ context.Context, _ model.Authentication) error {
					return errors.New("write failed")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "all generated keys collide — triggers loop-detected",
			method: http.MethodPost,
			body:   requestBody{Name: "Grace", Email: "grace@example.com"},
			store: &MockStore{
				CountApiPerUserFn: func(_ context.Context, _ string) (int, error) { return 0, nil },
				// Every key "already exists" → exhausts the retry loop
				ApiKeyExistsFn: func(_ context.Context, _ string) bool { return true },
			},
			wantStatusCode: http.StatusLoopDetected,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler(tc.store)

			var req *http.Request
			if s, ok := tc.body.(string); ok {
				// raw invalid JSON
				req = httptest.NewRequest(tc.method, "/envdash/v1/auth/", bytes.NewBufferString(s))
			} else {
				b, _ := json.Marshal(tc.body)
				req = httptest.NewRequest(tc.method, "/envdash/v1/auth/", bytes.NewReader(b))
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			h.Auth(rec, req)

			if rec.Code != tc.wantStatusCode {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatusCode, rec.Body.String())
			}

			if tc.wantKey {
				var resp map[string]string
				decodeBody(t, rec, &resp)
				if resp["key"] == "" {
					t.Error("expected non-empty 'key' in response body")
				}
				if resp["createdAt"] == "" {
					t.Error("expected non-empty 'createdAt' in response body")
				}
			}
		})
	}
}

// DeleteAuth

func TestDeleteAuth(t *testing.T) {
	const (
		validKey   = "sk-envdash-abc123"
		invalidKey = "sk-envdash-bad999"
	)

	tests := []struct {
		name           string
		method         string
		pathID         string // the key to delete (path param)
		headerKey      string // X-Api-Key header value
		store          *MockStore
		wantStatusCode int
	}{
		{
			name:      "happy path — delete own key",
			method:    http.MethodDelete,
			pathID:    validKey,
			headerKey: validKey,
			store: &MockStore{
				ApiKeyExistsFn: func(_ context.Context, _ string) bool { return true },
				DeleteAPIkeyFn: func(_ context.Context, _, _ string) error { return nil },
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:      "wrong method — POST not allowed",
			method:    http.MethodPost,
			pathID:    validKey,
			headerKey: validKey,
			store: &MockStore{
				ApiKeyExistsFn: func(_ context.Context, _ string) bool { return true },
			},
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:      "missing X-Api-Key header",
			method:    http.MethodDelete,
			pathID:    validKey,
			headerKey: "", // no header
			store: &MockStore{
				// empty string key → not found
				ApiKeyExistsFn: func(_ context.Context, _ string) bool { return false },
			},
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:      "header key does not exist in firestore",
			method:    http.MethodDelete,
			pathID:    validKey,
			headerKey: invalidKey,
			store: &MockStore{
				ApiKeyExistsFn: func(_ context.Context, _ string) bool { return false },
			},
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:      "target key not found in firestore",
			method:    http.MethodDelete,
			pathID:    "sk-envdash-nonexistent",
			headerKey: validKey,
			store: &MockStore{
				ApiKeyExistsFn: func(_ context.Context, _ string) bool { return true },
				DeleteAPIkeyFn: func(_ context.Context, _, _ string) error {
					return errors.New("api key not found")
				},
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:      "attempt to delete another user's key — unauthorized",
			method:    http.MethodDelete,
			pathID:    "sk-envdash-otheruser",
			headerKey: validKey,
			store: &MockStore{
				ApiKeyExistsFn: func(_ context.Context, _ string) bool { return true },
				DeleteAPIkeyFn: func(_ context.Context, _, _ string) error {
					return errors.New("unauthorized")
				},
			},
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:      "firestore error during deletion",
			method:    http.MethodDelete,
			pathID:    validKey,
			headerKey: validKey,
			store: &MockStore{
				ApiKeyExistsFn: func(_ context.Context, _ string) bool { return true },
				DeleteAPIkeyFn: func(_ context.Context, _, _ string) error {
					return errors.New("unexpected db error")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler(tc.store)

			req := httptest.NewRequest(tc.method, "/envdash/v1/auth/"+tc.pathID, nil)
			req.SetPathValue("id", tc.pathID)
			if tc.headerKey != "" {
				req.Header.Set("X-Api-Key", tc.headerKey)
			}

			rec := httptest.NewRecorder()
			h.Auth(rec, req)

			if rec.Code != tc.wantStatusCode {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatusCode, rec.Body.String())
			}

			// All error responses must include a JSON body with a non-empty "message" field
			if rec.Code >= 400 {
				var body model.ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
					t.Errorf("expected JSON error body, but could not decode: %v", err)
				} else if body.Message == "" {
					t.Error("expected non-empty 'message' field in error response body")
				}
			}
		})
	}
}

// Content-type header test

// Ensures that successful registration always returns application/json.
func TestRegisterAuth_ContentTypeHeader(t *testing.T) {
	store := &MockStore{
		CountApiPerUserFn:  func(_ context.Context, _ string) (int, error) { return 0, nil },
		ApiKeyExistsFn:     func(_ context.Context, _ string) bool { return false },
		CreateApiStorageFn: func(_ context.Context, _ model.Authentication) error { return nil },
	}

	h := newHandler(store)
	req := postAuthRequest(map[string]string{"name": "Tester", "email": "tester@example.com"})
	rec := httptest.NewRecorder()
	h.Auth(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

//
// API-key format test

// Ensures every generated key starts with the expected prefix.
func TestRegisterAuth_KeyPrefix(t *testing.T) {
	store := &MockStore{
		CountApiPerUserFn:  func(_ context.Context, _ string) (int, error) { return 0, nil },
		ApiKeyExistsFn:     func(_ context.Context, _ string) bool { return false },
		CreateApiStorageFn: func(_ context.Context, _ model.Authentication) error { return nil },
	}

	h := newHandler(store)
	req := postAuthRequest(map[string]string{"name": "Prefix Test", "email": "prefix@example.com"})
	rec := httptest.NewRecorder()
	h.Auth(rec, req)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)

	const prefix = "sk-envdash-"
	if len(resp["key"]) < len(prefix) || resp["key"][:len(prefix)] != prefix {
		t.Errorf("key %q does not start with %q", resp["key"], prefix)
	}
}

// Keys are unique

// Calls RegisterAuth multiple times and asserts all returned keys are unique.
func TestRegisterAuth_KeysAreUnique(t *testing.T) {
	seen := map[string]bool{}

	for i := 0; i < 20; i++ {
		store := &MockStore{
			CountApiPerUserFn:  func(_ context.Context, _ string) (int, error) { return 0, nil },
			ApiKeyExistsFn:     func(_ context.Context, _ string) bool { return false },
			CreateApiStorageFn: func(_ context.Context, _ model.Authentication) error { return nil },
		}
		h := newHandler(store)
		req := postAuthRequest(map[string]string{"name": "User", "email": "unique@example.com"})
		rec := httptest.NewRecorder()
		h.Auth(rec, req)

		var resp map[string]string
		json.NewDecoder(rec.Body).Decode(&resp)
		key := resp["key"]
		if seen[key] {
			t.Errorf("duplicate key generated: %s", key)
		}
		seen[key] = true
	}
}
