package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/testutil"
)

func createTestProfileHandler(mockRepo *MockProfileRepository) *ProfileHandler {
	return &ProfileHandler{
		Repo:      mockRepo,
		GetUserID: testutil.MockGetUserID,
		EncodeID:  testutil.MockEncodeID,
		DecodeID:  testutil.MockDecodeID,
	}
}

type MockProfileRepository struct {
	GetProfileByUserIDFunc func(ctx context.Context, userID string) (*model.Profile, error)
	CreateUserProfileFunc  func(ctx context.Context, profile *model.Profile, userID string) error
	UpdateProfileFunc      func(ctx context.Context, profile *model.Profile, userID string) error
}

func (m *MockProfileRepository) GetProfileByUserID(ctx context.Context, userID string) (*model.Profile, error) {
	if m.GetProfileByUserIDFunc != nil {
		return m.GetProfileByUserIDFunc(ctx, userID)
	}
	return nil, errors.New("GetProfileByUserID not implemented in mock")
}

func (m *MockProfileRepository) CreateUserProfile(ctx context.Context, profile *model.Profile, userID string) error {
	if m.CreateUserProfileFunc != nil {
		return m.CreateUserProfileFunc(ctx, profile, userID)
	}
	return errors.New("CreateUserProfile not implemented in mock")
}

func (m *MockProfileRepository) UpdateProfile(ctx context.Context, profile *model.Profile, userID string) error {
	if m.UpdateProfileFunc != nil {
		return m.UpdateProfileFunc(ctx, profile, userID)
	}
	return errors.New("UpdateProfile not implemented in mock")
}

func TestGetUserProfile_Success(t *testing.T) {
	mockRepo := &MockProfileRepository{
		GetProfileByUserIDFunc: func(ctx context.Context, userID string) (*model.Profile, error) {
			return testutil.MockProfile(), nil
		},
	}

	handler := createTestProfileHandler(mockRepo)
	req := testutil.CreateRequestWithUserID("GET", "/profile", "test")
	w := httptest.NewRecorder()

	handler.GetUserProfile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var profile model.Profile
	err := json.NewDecoder(w.Body).Decode(&profile)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if profile.ProfileID != "encoded" {
		t.Errorf("Expected ProfileID 'encoded', got '%s'", profile.ProfileID)
	}
}

func TestGetUserProfile_NoUserID(t *testing.T) {
	handler := createTestProfileHandler(&MockProfileRepository{})
	req := testutil.CreateRequestWithUserID("GET", "/profile", "")
	w := httptest.NewRecorder()

	handler.GetUserProfile(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetUserProfile_NotFound(t *testing.T) {
	mockRepo := &MockProfileRepository{
		GetProfileByUserIDFunc: func(ctx context.Context, userID string) (*model.Profile, error) {
			return nil, errors.New("profile not found")
		},
	}

	handler := createTestProfileHandler(mockRepo)
	req := testutil.CreateRequestWithUserID("GET", "/profile", "test")
	w := httptest.NewRecorder()

	handler.GetUserProfile(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestPostProfile_Success(t *testing.T) {
	mockRepo := &MockProfileRepository{
		CreateUserProfileFunc: func(ctx context.Context, profile *model.Profile, userID string) error {
			profile.ProfileID = "newProfileId"
			return nil
		},
	}

	handler := createTestProfileHandler(mockRepo)
	mockJSON := testutil.MockProfileJSON()
	req := testutil.CreateRequestWithBody("POST", "/profile", mockJSON, "test")
	w := httptest.NewRecorder()

	handler.PostUserProfile(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var profile model.Profile
	err := json.NewDecoder(w.Body).Decode(&profile)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if profile.ProfileID != "encoded" {
		t.Errorf("Expected ProfileID 'encoded', got '%s'", profile.ProfileID)
	}
}

func TestPostProfile_NoUserID(t *testing.T) {
	handler := createTestProfileHandler(&MockProfileRepository{})
	req := testutil.CreateRequestWithBody("POST", "/profile", "{}", "")
	w := httptest.NewRecorder()
	handler.PostUserProfile(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPostProfile_InvalidJSON(t *testing.T) {
	handler := createTestProfileHandler(&MockProfileRepository{})
	req := testutil.CreateRequestWithBody("POST", "/profile", "invalid json", "test")
	w := httptest.NewRecorder()
	handler.PostUserProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPostProfile_RepositoryError(t *testing.T) {
	mockRepo := &MockProfileRepository{
		CreateUserProfileFunc: func(ctx context.Context, profile *model.Profile, userID string) error {
			return errors.New("database error")
		},
	}

	handler := createTestProfileHandler(mockRepo)
	mockJSON := testutil.MockProfileJSON()
	req := testutil.CreateRequestWithBody("POST", "/profile", mockJSON, "test")
	w := httptest.NewRecorder()

	handler.PostUserProfile(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPutProfile_Success(t *testing.T) {
	mockRepo := &MockProfileRepository{
		UpdateProfileFunc: func(ctx context.Context, profile *model.Profile, userID string) error {
			return nil
		},
	}

	handler := createTestProfileHandler(mockRepo)
	mockJSON := testutil.MockProfileJSON()
	req := testutil.CreateRequestWithBody("PUT", "/profile", mockJSON, "test")
	w := httptest.NewRecorder()

	handler.PutUserProfile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var profile model.Profile
	err := json.NewDecoder(w.Body).Decode(&profile)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if profile.ProfileID != "profileId" {
		t.Errorf("Expected ProfileID = 'profileId', got '%s'", profile.ProfileID)
	}
}

func TestPutProfile_NoUserID(t *testing.T) {
	handler := createTestProfileHandler(&MockProfileRepository{})
	req := testutil.CreateRequestWithBody("PUT", "/profile", "{}", "")
	w := httptest.NewRecorder()

	handler.PutUserProfile(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPutProfile_InvalidJSON(t *testing.T) {
	handler := createTestProfileHandler(&MockProfileRepository{})
	req := testutil.CreateRequestWithBody("PUT", "/profile", "invalid json", "test")
	w := httptest.NewRecorder()

	handler.PutUserProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPutProfile_DecodeError(t *testing.T) {
	handler := createTestProfileHandler(&MockProfileRepository{})

	mockJSON := `{"profileId": "invalid"}`
	req := testutil.CreateRequestWithBody("PUT", "/profile", mockJSON, "test")
	w := httptest.NewRecorder()

	handler.PutUserProfile(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPutProfile_RepositoryError(t *testing.T) {
	mockRepo := &MockProfileRepository{
		UpdateProfileFunc: func(ctx context.Context, profile *model.Profile, userID string) error {
			return errors.New("update failed")
		},
	}

	handler := createTestProfileHandler(mockRepo)
	mockJSON := testutil.MockProfileJSON()
	req := testutil.CreateRequestWithBody("PUT", "/profile", mockJSON, "test")
	w := httptest.NewRecorder()

	handler.PutUserProfile(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
