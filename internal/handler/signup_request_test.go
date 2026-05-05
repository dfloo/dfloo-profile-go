package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/testutil"
)

type MockSignupRequestRepository struct {
	CreateSignupRequestFunc func(ctx context.Context, req *model.SignupRequest) error
}

func (m *MockSignupRequestRepository) CreateSignupRequest(ctx context.Context, req *model.SignupRequest) error {
	if m.CreateSignupRequestFunc != nil {
		return m.CreateSignupRequestFunc(ctx, req)
	}
	return errors.New("CreateSignupRequest not implemented in mock")
}

func createTestSignupRequestHandler(mockRepo *MockSignupRequestRepository, sendEmail func(string, string) error) *SignupRequestHandler {
	return &SignupRequestHandler{Repo: mockRepo, SendEmail: sendEmail}
}

func TestPostSignupRequest_Success(t *testing.T) {
	mockRepo := &MockSignupRequestRepository{
		CreateSignupRequestFunc: func(ctx context.Context, req *model.SignupRequest) error { return nil },
	}

	handler := createTestSignupRequestHandler(mockRepo, noopSendEmail)
	body := `{"name":"Bob","email":"bob@example.com","reason":"I want to join"}`
	req := testutil.CreateRequestWithBody("POST", "/api/signup-requests", body, "")
	w := httptest.NewRecorder()

	handler.PostSignupRequest(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestPostSignupRequest_MissingFields(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"missing name", `{"email":"b@c.com","reason":"hi"}`},
		{"missing email", `{"name":"Bob","reason":"hi"}`},
		{"missing reason", `{"name":"Bob","email":"b@c.com"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := createTestSignupRequestHandler(&MockSignupRequestRepository{}, noopSendEmail)
			req := testutil.CreateRequestWithBody("POST", "/api/signup-requests", tc.body, "")
			w := httptest.NewRecorder()

			handler.PostSignupRequest(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestPostSignupRequest_InvalidJSON(t *testing.T) {
	handler := createTestSignupRequestHandler(&MockSignupRequestRepository{}, noopSendEmail)
	req := testutil.CreateRequestWithBody("POST", "/api/signup-requests", "not json", "")
	w := httptest.NewRecorder()

	handler.PostSignupRequest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPostSignupRequest_RepositoryError(t *testing.T) {
	mockRepo := &MockSignupRequestRepository{
		CreateSignupRequestFunc: func(ctx context.Context, req *model.SignupRequest) error {
			return errors.New("db error")
		},
	}

	handler := createTestSignupRequestHandler(mockRepo, noopSendEmail)
	body := `{"name":"Bob","email":"bob@example.com","reason":"I want to join"}`
	req := testutil.CreateRequestWithBody("POST", "/api/signup-requests", body, "")
	w := httptest.NewRecorder()

	handler.PostSignupRequest(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPostSignupRequest_EmailFailureDoesNotAffectResponse(t *testing.T) {
	mockRepo := &MockSignupRequestRepository{
		CreateSignupRequestFunc: func(ctx context.Context, req *model.SignupRequest) error { return nil },
	}

	handler := createTestSignupRequestHandler(mockRepo, func(subject, html string) error {
		return errors.New("smtp down")
	})
	body := `{"name":"Bob","email":"bob@example.com","reason":"I want to join"}`
	req := testutil.CreateRequestWithBody("POST", "/api/signup-requests", body, "")
	w := httptest.NewRecorder()

	handler.PostSignupRequest(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestPostSignupRequest_EmailSubjectContainsName(t *testing.T) {
	mockRepo := &MockSignupRequestRepository{
		CreateSignupRequestFunc: func(ctx context.Context, req *model.SignupRequest) error { return nil },
	}

	var capturedSubject string
	handler := createTestSignupRequestHandler(mockRepo, func(subject, html string) error {
		capturedSubject = subject
		return nil
	})
	body := `{"name":"Bob","email":"bob@example.com","reason":"I want to join"}`
	req := testutil.CreateRequestWithBody("POST", "/api/signup-requests", body, "")
	w := httptest.NewRecorder()

	handler.PostSignupRequest(w, req)

	if !strings.Contains(capturedSubject, "Bob") {
		t.Errorf("expected subject to contain sender name, got %q", capturedSubject)
	}
}
