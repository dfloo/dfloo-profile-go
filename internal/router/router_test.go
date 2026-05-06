package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_ProtectedRoutesRequireAuthentication(t *testing.T) {
	t.Setenv("AUTH0_DOMAIN", "example.auth0.com")
	t.Setenv("AUTH0_AUDIENCE", "https://api.example.com")
	t.Setenv("CLIENT_ORIGIN", "http://localhost:3000")

	mux := New(nil)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "profiles", method: http.MethodGet, path: "/api/profiles"},
		{name: "resumes", method: http.MethodGet, path: "/api/resumes"},
		{name: "job applications", method: http.MethodGet, path: "/api/job-applications"},
		{name: "set default resume", method: http.MethodPost, path: "/api/resumes/default"},
		{name: "download resume", method: http.MethodPost, path: "/api/resumes/download/encoded"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestNew_PublicDownloadRouteHandlesPreflight(t *testing.T) {
	t.Setenv("CLIENT_ORIGIN", "http://localhost:3000")

	mux := New(nil)
	req := httptest.NewRequest(http.MethodOptions, "/api/download/resume/default", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
	}
}

func TestNew_F1ChampionshipsPlaceholderHandler(t *testing.T) {
	t.Setenv("CLIENT_ORIGIN", "http://localhost:3000")

	mux := New(nil)

	t.Run("GET returns internal server error without database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/f1/championships", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}

		var payload map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("response body is not valid JSON object: %v", err)
		}
		if payload["message"] == "" {
			t.Fatalf("expected non-empty message field")
		}
	})

	t.Run("non-GET methods return method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/f1/championships", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestNew_F1DriversHandler(t *testing.T) {
	t.Setenv("CLIENT_ORIGIN", "http://localhost:3000")

	mux := New(nil)

	t.Run("GET with missing year returns bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}

		var payload map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("response body is not valid JSON object: %v", err)
		}
		if payload["message"] == "" {
			t.Fatalf("expected non-empty message field")
		}
	})

	t.Run("non-GET methods return method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/f1/drivers", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestNew_F1DriverDetailsHandler(t *testing.T) {
	t.Setenv("CLIENT_ORIGIN", "http://localhost:3000")

	mux := New(nil)

	t.Run("GET with missing year returns bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers/1", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}

		var payload map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("response body is not valid JSON object: %v", err)
		}
		if payload["message"] == "" {
			t.Fatalf("expected non-empty message field")
		}
	})

	t.Run("non-GET methods return method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/f1/drivers/1?year=2024", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestNew_F1ConstructorsHandler(t *testing.T) {
	t.Setenv("CLIENT_ORIGIN", "http://localhost:3000")

	mux := New(nil)

	t.Run("GET with missing year returns bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/f1/constructors", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}

		var payload map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("response body is not valid JSON object: %v", err)
		}
		if payload["message"] == "" {
			t.Fatalf("expected non-empty message field")
		}
	})

	t.Run("non-GET methods return method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/f1/constructors", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestNew_F1EventsHandler(t *testing.T) {
	t.Setenv("CLIENT_ORIGIN", "http://localhost:3000")

	mux := New(nil)

	t.Run("GET with missing year returns bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/f1/events", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}

		var payload map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("response body is not valid JSON object: %v", err)
		}
		if payload["message"] == "" {
			t.Fatalf("expected non-empty message field")
		}
	})

	t.Run("GET with year returns internal server error without database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/f1/events?year=2024", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}

		var payload map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("response body is not valid JSON object: %v", err)
		}
		if payload["message"] == "" {
			t.Fatalf("expected non-empty message field")
		}
	})

	t.Run("non-GET methods return method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/f1/events", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestNew_FontOptionsHandler(t *testing.T) {
	t.Setenv("CLIENT_ORIGIN", "http://localhost:3000")

	mux := New(nil)

	t.Run("GET returns 200 with font options", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/resumes/fonts", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
		if ct := w.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type = %q, want %q", ct, "application/json")
		}
		var opts []map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &opts); err != nil {
			t.Fatalf("body is not valid JSON array: %v", err)
		}
		if len(opts) == 0 {
			t.Fatalf("expected non-empty font options")
		}
	})

	t.Run("non-GET methods return method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/resumes/fonts", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestNew_ProtectedDownloadRouteHandlesPreflight(t *testing.T) {
	t.Setenv("AUTH0_DOMAIN", "example.auth0.com")
	t.Setenv("AUTH0_AUDIENCE", "https://api.example.com")
	t.Setenv("CLIENT_ORIGIN", "http://localhost:3000")

	mux := New(nil)
	req := httptest.NewRequest(http.MethodOptions, "/api/resumes/download/encoded", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
	}
}
