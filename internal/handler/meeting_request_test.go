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

type MockMeetingRequestRepository struct {
	CreateMeetingRequestFunc func(ctx context.Context, req *model.MeetingRequest) error
}

func (m *MockMeetingRequestRepository) CreateMeetingRequest(ctx context.Context, req *model.MeetingRequest) error {
	if m.CreateMeetingRequestFunc != nil {
		return m.CreateMeetingRequestFunc(ctx, req)
	}
	return errors.New("CreateMeetingRequest not implemented in mock")
}

func noopSendEmail(subject, html string) error { return nil }

func createTestMeetingRequestHandler(mockRepo *MockMeetingRequestRepository, sendEmail func(string, string) error) *MeetingRequestHandler {
	return &MeetingRequestHandler{Repo: mockRepo, SendEmail: sendEmail}
}

func TestPostMeetingRequest_Success(t *testing.T) {
	mockRepo := &MockMeetingRequestRepository{
		CreateMeetingRequestFunc: func(ctx context.Context, req *model.MeetingRequest) error { return nil },
	}

	handler := createTestMeetingRequestHandler(mockRepo, noopSendEmail)
	body := `{"name":"Alice","email":"alice@example.com","message":"Hello"}`
	req := testutil.CreateRequestWithBody("POST", "/api/meeting-requests", body, "")
	w := httptest.NewRecorder()

	handler.PostMeetingRequest(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestPostMeetingRequest_MissingFields(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"missing name", `{"email":"a@b.com","message":"hi"}`},
		{"missing email", `{"name":"Alice","message":"hi"}`},
		{"missing message", `{"name":"Alice","email":"a@b.com"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := createTestMeetingRequestHandler(&MockMeetingRequestRepository{}, noopSendEmail)
			req := testutil.CreateRequestWithBody("POST", "/api/meeting-requests", tc.body, "")
			w := httptest.NewRecorder()

			handler.PostMeetingRequest(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestPostMeetingRequest_InvalidJSON(t *testing.T) {
	handler := createTestMeetingRequestHandler(&MockMeetingRequestRepository{}, noopSendEmail)
	req := testutil.CreateRequestWithBody("POST", "/api/meeting-requests", "not json", "")
	w := httptest.NewRecorder()

	handler.PostMeetingRequest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPostMeetingRequest_RepositoryError(t *testing.T) {
	mockRepo := &MockMeetingRequestRepository{
		CreateMeetingRequestFunc: func(ctx context.Context, req *model.MeetingRequest) error {
			return errors.New("db error")
		},
	}

	handler := createTestMeetingRequestHandler(mockRepo, noopSendEmail)
	body := `{"name":"Alice","email":"alice@example.com","message":"Hello"}`
	req := testutil.CreateRequestWithBody("POST", "/api/meeting-requests", body, "")
	w := httptest.NewRecorder()

	handler.PostMeetingRequest(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPostMeetingRequest_EmailFailureDoesNotAffectResponse(t *testing.T) {
	mockRepo := &MockMeetingRequestRepository{
		CreateMeetingRequestFunc: func(ctx context.Context, req *model.MeetingRequest) error { return nil },
	}

	handler := createTestMeetingRequestHandler(mockRepo, func(subject, html string) error {
		return errors.New("smtp down")
	})
	body := `{"name":"Alice","email":"alice@example.com","message":"Hello"}`
	req := testutil.CreateRequestWithBody("POST", "/api/meeting-requests", body, "")
	w := httptest.NewRecorder()

	handler.PostMeetingRequest(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestPostMeetingRequest_EmailSubjectContainsName(t *testing.T) {
	mockRepo := &MockMeetingRequestRepository{
		CreateMeetingRequestFunc: func(ctx context.Context, req *model.MeetingRequest) error { return nil },
	}

	var capturedSubject string
	handler := createTestMeetingRequestHandler(mockRepo, func(subject, html string) error {
		capturedSubject = subject
		return nil
	})
	body := `{"name":"Alice","email":"alice@example.com","message":"Hello"}`
	req := testutil.CreateRequestWithBody("POST", "/api/meeting-requests", body, "")
	w := httptest.NewRecorder()

	handler.PostMeetingRequest(w, req)

	if !strings.Contains(capturedSubject, "Alice") {
		t.Errorf("expected subject to contain sender name, got %q", capturedSubject)
	}
}
