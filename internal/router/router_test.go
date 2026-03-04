package router

import (
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

func TestNew_PublicMetricsRoute(t *testing.T) {
	mux := New(nil)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Content-Type"); got == "" {
		t.Fatal("Content-Type header is empty")
	}
}
