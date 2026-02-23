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

func createTestResumeHandler(mockRepo *MockResumeRepository) *ResumeHandler {
	return &ResumeHandler{
		Repo:      mockRepo,
		GetUserID: testutil.MockGetUserID,
		EncodeID:  testutil.MockEncodeID,
		DecodeID:  testutil.MockDecodeID,
	}
}

type MockResumeRepository struct {
	GetResumesByUserIDFunc func(ctx context.Context, userID string) ([]*model.Resume, error)
	GetResumeByIDFunc      func(ctx context.Context, resumeID, userID string) (*model.Resume, error)
	GetDefaultResumeFunc   func(ctx context.Context) (*model.Resume, error)
	CreateResumeFunc       func(ctx context.Context, resume *model.Resume, userID string) error
	UpdateResumeFunc       func(ctx context.Context, resume *model.Resume, userID string) error
	DeleteResumesFunc      func(ctx context.Context, resumeIDs []string, userID string) ([]string, error)
}

func (m *MockResumeRepository) GetResumesByUserID(ctx context.Context, userID string) ([]*model.Resume, error) {
	if m.GetResumesByUserIDFunc != nil {
		return m.GetResumesByUserIDFunc(ctx, userID)
	}
	return []*model.Resume{}, errors.New("GetResumesByUserID not implemented in mock")
}

func (m *MockResumeRepository) GetResumeByID(ctx context.Context, resumeID, userID string) (*model.Resume, error) {
	if m.GetResumeByIDFunc != nil {
		return m.GetResumeByIDFunc(ctx, resumeID, userID)
	}
	return nil, errors.New("GetResumeByID not implemented in mock")
}

func (m *MockResumeRepository) GetDefaultResume(ctx context.Context) (*model.Resume, error) {
	if m.GetDefaultResumeFunc != nil {
		return m.GetDefaultResumeFunc(ctx)
	}
	return nil, errors.New("GetDefaultResume not implemented in mock")
}

func (m *MockResumeRepository) CreateResume(ctx context.Context, resume *model.Resume, userID string) error {
	if m.CreateResumeFunc != nil {
		return m.CreateResumeFunc(ctx, resume, userID)
	}
	return errors.New("CreateResume not implemented in mock")
}

func (m *MockResumeRepository) UpdateResume(ctx context.Context, resume *model.Resume, userID string) error {
	if m.UpdateResumeFunc != nil {
		return m.UpdateResumeFunc(ctx, resume, userID)
	}
	return errors.New("UpdateResume not implemented in mock")
}

func (m *MockResumeRepository) DeleteResumes(ctx context.Context, resumeIDs []string, userID string) ([]string, error) {
	if m.DeleteResumesFunc != nil {
		return m.DeleteResumesFunc(ctx, resumeIDs, userID)
	}
	return nil, errors.New("DeleteResumes not implemented in mock")
}

func TestGetUserResumes_Success(t *testing.T) {
	mockRepo := &MockResumeRepository{
		GetResumesByUserIDFunc: func(ctx context.Context, userID string) ([]*model.Resume, error) {
			return []*model.Resume{testutil.MockResume()}, nil
		},
	}

	handler := createTestResumeHandler(mockRepo)
	req := testutil.CreateRequestWithUserID("GET", "/resumes", "test")
	w := httptest.NewRecorder()

	handler.GetUserResumes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resumes []*model.Resume
	err := json.NewDecoder(w.Body).Decode(&resumes)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(resumes) != 1 {
		t.Errorf("Expected 1 resume, got %d", len(resumes))
	}

	if resumes[0].ResumeID != "encoded" {
		t.Errorf("Expected ResumeID 'encoded', got '%s'", resumes[0].ResumeID)
	}

	if resumes[0].Profile.ProfileID != "encoded" {
		t.Errorf("Expected ProfileID 'encoded', got '%s'", resumes[0].Profile.ProfileID)
	}
}

func TestGetUserResumes_NoUserID(t *testing.T) {
	handler := createTestResumeHandler(&MockResumeRepository{})
	req := testutil.CreateRequestWithUserID("GET", "/resumes", "")
	w := httptest.NewRecorder()

	handler.GetUserResumes(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetUserResumes_NotFound(t *testing.T) {
	mockRepo := &MockResumeRepository{
		GetResumesByUserIDFunc: func(ctx context.Context, userID string) ([]*model.Resume, error) {
			return []*model.Resume{}, errors.New("database error")
		},
	}

	handler := createTestResumeHandler(mockRepo)
	req := testutil.CreateRequestWithUserID("GET", "/resumes", "test")
	w := httptest.NewRecorder()

	handler.GetUserResumes(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestPostResume_Success(t *testing.T) {
	mockRepo := &MockResumeRepository{
		CreateResumeFunc: func(ctx context.Context, resume *model.Resume, userID string) error {
			resume.ResumeID = "newResumeId"
			resume.Profile.ProfileID = "newProfileId"
			return nil
		},
	}

	resumeJSON := testutil.MockResumeJSON()
	handler := createTestResumeHandler(mockRepo)
	req := testutil.CreateRequestWithBody("POST", "/resumes", resumeJSON, "test")
	w := httptest.NewRecorder()

	handler.PostResume(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resume model.Resume
	err := json.NewDecoder(w.Body).Decode(&resume)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if resume.ResumeID != "encoded" {
		t.Errorf("Expected ResumeID 'encoded', got '%s'", resume.ResumeID)
	}

	if resume.Profile.ProfileID != "encoded" {
		t.Errorf("Expected ProfileID 'encoded', got '%s'", resume.Profile.ProfileID)
	}
}

func TestPostResume_NoUserID(t *testing.T) {
	handler := createTestResumeHandler(&MockResumeRepository{})
	req := testutil.CreateRequestWithBody("POST", "/resumes", "{}", "")
	w := httptest.NewRecorder()

	handler.PostResume(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPostResume_InvalidJSON(t *testing.T) {
	handler := createTestResumeHandler(&MockResumeRepository{})
	req := testutil.CreateRequestWithBody("POST", "/resumes", "invalid json", "test")
	w := httptest.NewRecorder()

	handler.PostResume(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPostResume_RepositoryError(t *testing.T) {
	mockRepo := &MockResumeRepository{
		CreateResumeFunc: func(ctx context.Context, resume *model.Resume, userID string) error {
			return errors.New("database error")
		},
	}

	handler := createTestResumeHandler(mockRepo)
	req := testutil.CreateRequestWithBody("POST", "/resumes", "{}", "test")
	w := httptest.NewRecorder()

	handler.PostResume(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPutResume_Success(t *testing.T) {
	mockRepo := &MockResumeRepository{
		UpdateResumeFunc: func(ctx context.Context, resume *model.Resume, userID string) error {
			return nil
		},
	}

	resumeJSON := testutil.MockResumeJSON()
	handler := createTestResumeHandler(mockRepo)
	req := testutil.CreateRequestWithBody("PUT", "/resumes", resumeJSON, "test")
	w := httptest.NewRecorder()

	handler.PutResume(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resume model.Resume
	err := json.NewDecoder(w.Body).Decode(&resume)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if resume.ResumeID != "resumeId" {
		t.Errorf("Expected encoded ResumeID 'resumeId', got '%s'", resume.ResumeID)
	}
}

func TestPutResume_NoUserID(t *testing.T) {
	handler := createTestResumeHandler(&MockResumeRepository{})
	req := testutil.CreateRequestWithBody("PUT", "/resumes", "{}", "")
	w := httptest.NewRecorder()

	handler.PutResume(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPutResume_InvalidJSON(t *testing.T) {
	handler := createTestResumeHandler(&MockResumeRepository{})
	req := testutil.CreateRequestWithBody("PUT", "/resumes", "invalid json", "test")
	w := httptest.NewRecorder()

	handler.PutResume(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPutResume_DecodeError(t *testing.T) {
	handler := createTestResumeHandler(&MockResumeRepository{})

	resumeJSON := `{"resumeId": "invalid"}`
	req := testutil.CreateRequestWithBody("PUT", "/resumes", resumeJSON, "test")
	w := httptest.NewRecorder()

	handler.PutResume(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPutResume_RepositoryError(t *testing.T) {
	mockRepo := &MockResumeRepository{
		UpdateResumeFunc: func(ctx context.Context, resume *model.Resume, userID string) error {
			return errors.New("update failed")
		},
	}

	handler := createTestResumeHandler(mockRepo)

	resumeJSON := testutil.MockResumeJSON()
	req := testutil.CreateRequestWithBody("PUT", "/resumes", resumeJSON, "test")
	w := httptest.NewRecorder()

	handler.PutResume(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDeleteResumes_Success(t *testing.T) {
	mockRepo := &MockResumeRepository{
		DeleteResumesFunc: func(ctx context.Context, resumeIDs []string, userID string) ([]string, error) {
			expectedDecodedIDs := []string{"decoded", "decoded"}
			if len(resumeIDs) != len(expectedDecodedIDs) {
				return nil, errors.New("unexpected number of IDs")
			}
			return resumeIDs, nil
		},
	}

	resumeIDsJSON := `["resume1", "resume2"]`
	handler := createTestResumeHandler(mockRepo)
	req := testutil.CreateRequestWithBody("DELETE", "/resumes", resumeIDsJSON, "test")
	w := httptest.NewRecorder()

	handler.DeleteResumes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var deletedIDs []string
	err := json.NewDecoder(w.Body).Decode(&deletedIDs)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(deletedIDs) != 2 {
		t.Errorf("Expected 2 deleted IDs, got %d", len(deletedIDs))
	}

	for _, id := range deletedIDs {
		if id != "encoded" {
			t.Errorf("Expected ID 'encoded', got '%s'", id)
		}
	}
}

func TestDeleteResumes_PartialSuccess(t *testing.T) {
	mockRepo := &MockResumeRepository{
		DeleteResumesFunc: func(ctx context.Context, resumeIDs []string, userID string) ([]string, error) {
			if len(resumeIDs) >= 1 {
				return []string{resumeIDs[0]}, nil
			}
			return []string{}, nil
		},
	}

	resumeIDsJSON := `["resume1", "resume2"]`
	handler := createTestResumeHandler(mockRepo)
	req := testutil.CreateRequestWithBody("DELETE", "/resumes", resumeIDsJSON, "test")
	w := httptest.NewRecorder()

	handler.DeleteResumes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var deletedIDs []string
	err := json.NewDecoder(w.Body).Decode(&deletedIDs)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(deletedIDs) != 1 {
		t.Errorf("Expected 1 deleted ID, got %d", len(deletedIDs))
	}

	if deletedIDs[0] != "encoded" {
		t.Errorf("Expected ID 'encoded', got '%s'", deletedIDs[0])
	}
}

func TestDeleteResumes_NoUserID(t *testing.T) {
	handler := createTestResumeHandler(&MockResumeRepository{})
	req := testutil.CreateRequestWithBody("DELETE", "/resumes", `["resume1"]`, "")
	w := httptest.NewRecorder()

	handler.DeleteResumes(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestDeleteResumes_InvalidJSON(t *testing.T) {
	handler := createTestResumeHandler(&MockResumeRepository{})
	req := testutil.CreateRequestWithBody("DELETE", "/resumes", "invalid json", "test")
	w := httptest.NewRecorder()

	handler.DeleteResumes(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteResumes_DecodeError(t *testing.T) {
	handler := createTestResumeHandler(&MockResumeRepository{})

	req := testutil.CreateRequestWithBody("DELETE", "/resumes", `["invalid"]`, "test")
	w := httptest.NewRecorder()

	handler.DeleteResumes(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDeleteResumes_RepositoryError(t *testing.T) {
	mockRepo := &MockResumeRepository{
		DeleteResumesFunc: func(ctx context.Context, resumeIDs []string, userID string) ([]string, error) {
			return nil, errors.New("delete failed")
		},
	}

	handler := createTestResumeHandler(mockRepo)
	req := testutil.CreateRequestWithBody("DELETE", "/resumes", `["resume1"]`, "test")
	w := httptest.NewRecorder()

	handler.DeleteResumes(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDownloadResumePDF_InvalidJSON(t *testing.T) {
	handler := createTestResumeHandler(&MockResumeRepository{})
	req := testutil.CreateRequestWithBody("POST", "/download", "invalid json", "")
	w := httptest.NewRecorder()

	handler.DownloadResumePDF(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
