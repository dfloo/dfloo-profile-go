package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

type otherClaims struct{}

func (otherClaims) Validate(context.Context) error {
	return nil
}

func TestCustomClaimsHasScope(t *testing.T) {
	tests := []struct {
		name     string
		scope    string
		expected string
		want     bool
	}{
		{name: "scope present", scope: "read:profile write:profile", expected: "write:profile", want: true},
		{name: "scope absent", scope: "read:profile", expected: "write:profile", want: false},
		{name: "empty scope", scope: "", expected: "write:profile", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims := CustomClaims{Scope: tc.scope}
			got := claims.HasScope(tc.expected)
			if got != tc.want {
				t.Errorf("HasScope(%q) = %v, want %v", tc.expected, got, tc.want)
			}
		})
	}
}

func TestGetPermissions(t *testing.T) {
	validClaims := &validator.ValidatedClaims{
		CustomClaims: &CustomClaims{Permissions: []string{"resume:read", "resume:write"}},
	}

	tests := []struct {
		name string
		ctx  context.Context
		want []string
	}{
		{name: "missing claims", ctx: context.Background(), want: nil},
		{name: "wrong claims type", ctx: context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, "invalid"), want: nil},
		{name: "wrong custom claims type", ctx: context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: otherClaims{}}), want: nil},
		{name: "success", ctx: context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, validClaims), want: []string{"resume:read", "resume:write"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetPermissions(tc.ctx)
			if len(got) != len(tc.want) {
				t.Fatalf("GetPermissions() len = %d, want %d", len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("GetPermissions()[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestHasPermission(t *testing.T) {
	claims := &validator.ValidatedClaims{
		CustomClaims: &CustomClaims{Permissions: []string{"job:read", "job:write"}},
	}
	ctx := context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, claims)

	tests := []struct {
		name       string
		permission string
		want       bool
	}{
		{name: "permission exists", permission: "job:write", want: true},
		{name: "permission missing", permission: "profile:read", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := HasPermission(ctx, tc.permission)
			if got != tc.want {
				t.Errorf("HasPermission(%q) = %v, want %v", tc.permission, got, tc.want)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{name: "missing claims", ctx: context.Background(), want: ""},
		{name: "wrong claims type", ctx: context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, 123), want: ""},
		{
			name: "success",
			ctx: context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{
				RegisteredClaims: validator.RegisteredClaims{Subject: "auth0|user-123"},
			}),
			want: "auth0|user-123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetUserID(tc.ctx)
			if got != tc.want {
				t.Errorf("GetUserID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEncodeDecodeID(t *testing.T) {
	decodedInput := "resume-id-123"
	encoded := EncodeID(decodedInput)

	decoded, err := DecodeID(encoded)
	if err != nil {
		t.Fatalf("DecodeID() returned unexpected error: %v", err)
	}
	if decoded != decodedInput {
		t.Fatalf("DecodeID(EncodeID(input)) = %q, want %q", decoded, decodedInput)
	}

	_, err = DecodeID("%%%")
	if err == nil {
		t.Fatalf("DecodeID() expected error for invalid base64 input")
	}
}

func TestCORS(t *testing.T) {
	t.Setenv("CLIENT_ORIGIN", "http://localhost:3000")

	t.Run("options short circuit", func(t *testing.T) {
		called := false
		h := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusTeapot)
		}))

		req := httptest.NewRequest(http.MethodOptions, "/api/profiles", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if called {
			t.Fatalf("expected next handler not to be called for OPTIONS requests")
		}
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
			t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
		}
		if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET,POST,PUT,DELETE,OPTIONS" {
			t.Fatalf("Access-Control-Allow-Methods = %q, want %q", got, "GET,POST,PUT,DELETE,OPTIONS")
		}
	})

	t.Run("non options calls next", func(t *testing.T) {
		called := false
		h := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusNoContent)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/profiles", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if !called {
			t.Fatalf("expected next handler to be called for non-OPTIONS requests")
		}
		if w.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
		}
		if got := w.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type" {
			t.Fatalf("Access-Control-Allow-Headers = %q, want %q", got, "Authorization, Content-Type")
		}
	})
}
