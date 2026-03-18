package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/testutil"
)

func createTestResumeHandler(t *testing.T, mockRepo *MockResumeRepository) *ResumeHandler {
	t.Helper()
	cacheDir := t.TempDir()
	texDir := t.TempDir()

	return &ResumeHandler{
		Repo:      mockRepo,
		GetUserID: testutil.MockGetUserID,
		EncodeID:  testutil.MockEncodeID,
		DecodeID:  testutil.MockDecodeID,
		CacheDir:  cacheDir,
		GenerateFromResume: func(resume *model.Resume) (string, error) {
			return filepath.Join(texDir, "resume.tex"), nil
		},
		ConvertToPDF: func(filePath string) ([]byte, error) {
			return []byte("test-pdf"), nil
		},
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

	handler := createTestResumeHandler(t, mockRepo)
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
	handler := createTestResumeHandler(t, &MockResumeRepository{})
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

	handler := createTestResumeHandler(t, mockRepo)
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
	handler := createTestResumeHandler(t, mockRepo)
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
	handler := createTestResumeHandler(t, &MockResumeRepository{})
	req := testutil.CreateRequestWithBody("POST", "/resumes", "{}", "")
	w := httptest.NewRecorder()

	handler.PostResume(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPostResume_InvalidJSON(t *testing.T) {
	handler := createTestResumeHandler(t, &MockResumeRepository{})
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

	handler := createTestResumeHandler(t, mockRepo)
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
	handler := createTestResumeHandler(t, mockRepo)
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
	handler := createTestResumeHandler(t, &MockResumeRepository{})
	req := testutil.CreateRequestWithBody("PUT", "/resumes", "{}", "")
	w := httptest.NewRecorder()

	handler.PutResume(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPutResume_InvalidJSON(t *testing.T) {
	handler := createTestResumeHandler(t, &MockResumeRepository{})
	req := testutil.CreateRequestWithBody("PUT", "/resumes", "invalid json", "test")
	w := httptest.NewRecorder()

	handler.PutResume(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPutResume_DecodeError(t *testing.T) {
	handler := createTestResumeHandler(t, &MockResumeRepository{})

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

	handler := createTestResumeHandler(t, mockRepo)

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
	handler := createTestResumeHandler(t, mockRepo)
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

func TestDeleteResumes_RemovesCacheFiles(t *testing.T) {
	deletedDecodedID := "decoded"

	mockRepo := &MockResumeRepository{
		DeleteResumesFunc: func(ctx context.Context, resumeIDs []string, userID string) ([]string, error) {
			if len(resumeIDs) != 1 || resumeIDs[0] != deletedDecodedID {
				return nil, errors.New("unexpected resume IDs")
			}
			return resumeIDs, nil
		},
	}

	handler := createTestResumeHandler(t, mockRepo)
	cacheFilePath := handler.resumeCacheFilePath(deletedDecodedID)
	if err := os.WriteFile(cacheFilePath, []byte("cached-pdf"), 0644); err != nil {
		t.Fatalf("Failed to write cache file: %v", err)
	}

	req := testutil.CreateRequestWithBody("DELETE", "/resumes", `["resume1"]`, "test")
	w := httptest.NewRecorder()

	handler.DeleteResumes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if _, err := os.Stat(cacheFilePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Expected cache file to be removed, stat err=%v", err)
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
	handler := createTestResumeHandler(t, mockRepo)
	deletedCacheFilePath := handler.resumeCacheFilePath("decoded")
	if err := os.WriteFile(deletedCacheFilePath, []byte("cached-pdf"), 0644); err != nil {
		t.Fatalf("Failed to write deleted cache file: %v", err)
	}
	notDeletedDecodedID := "other-resume"
	notDeletedCacheFilePath := handler.resumeCacheFilePath(notDeletedDecodedID)
	if err := os.WriteFile(notDeletedCacheFilePath, []byte("cached-pdf"), 0644); err != nil {
		t.Fatalf("Failed to write non-deleted cache file: %v", err)
	}
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

	if _, err := os.Stat(deletedCacheFilePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Expected deleted cache file to be removed, stat err=%v", err)
	}

	if _, err := os.Stat(notDeletedCacheFilePath); err != nil {
		t.Fatalf("Expected non-deleted cache file to remain, stat err=%v", err)
	}
}

func TestDeleteResumes_NoUserID(t *testing.T) {
	handler := createTestResumeHandler(t, &MockResumeRepository{})
	req := testutil.CreateRequestWithBody("DELETE", "/resumes", `["resume1"]`, "")
	w := httptest.NewRecorder()

	handler.DeleteResumes(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestDeleteResumes_InvalidJSON(t *testing.T) {
	handler := createTestResumeHandler(t, &MockResumeRepository{})
	req := testutil.CreateRequestWithBody("DELETE", "/resumes", "invalid json", "test")
	w := httptest.NewRecorder()

	handler.DeleteResumes(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteResumes_DecodeError(t *testing.T) {
	handler := createTestResumeHandler(t, &MockResumeRepository{})

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

	handler := createTestResumeHandler(t, mockRepo)
	req := testutil.CreateRequestWithBody("DELETE", "/resumes", `["resume1"]`, "test")
	w := httptest.NewRecorder()

	handler.DeleteResumes(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDownloadResumePDF_NoUserID(t *testing.T) {
	handler := createTestResumeHandler(t, &MockResumeRepository{})
	req := testutil.CreateRequestWithUserID("GET", "/download/encoded", "")
	req.SetPathValue("resumeId", "encoded")
	w := httptest.NewRecorder()

	handler.DownloadResumePDF(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestDownloadResumePDF_MissingResumeID(t *testing.T) {
	handler := createTestResumeHandler(t, &MockResumeRepository{})
	handler.CacheDir = t.TempDir()
	req := testutil.CreateRequestWithUserID("GET", "/download", "test")
	w := httptest.NewRecorder()

	handler.DownloadResumePDF(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDownloadResumePDF_DecodeError(t *testing.T) {
	handler := createTestResumeHandler(t, &MockResumeRepository{})
	req := testutil.CreateRequestWithUserID("GET", "/download/invalid", "test")
	req.SetPathValue("resumeId", "invalid")
	w := httptest.NewRecorder()

	handler.DownloadResumePDF(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDownloadResumePDF_ResumeNotFound(t *testing.T) {
	mockRepo := &MockResumeRepository{
		GetResumeByIDFunc: func(ctx context.Context, resumeID, userID string) (*model.Resume, error) {
			return nil, errors.New("not found")
		},
	}

	handler := createTestResumeHandler(t, mockRepo)
	req := testutil.CreateRequestWithUserID("GET", "/download/encoded", "test")
	req.SetPathValue("resumeId", "encoded")
	w := httptest.NewRecorder()

	handler.DownloadResumePDF(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDownloadResumePDF_StartsGenerationOnCacheMiss(t *testing.T) {
	mockRepo := &MockResumeRepository{
		GetResumeByIDFunc: func(ctx context.Context, resumeID, userID string) (*model.Resume, error) {
			if userID != "test" {
				t.Fatalf("Expected userID test, got %s", userID)
			}
			if resumeID != "decoded" {
				t.Fatalf("Expected resumeID decoded, got %s", resumeID)
			}
			resume := testutil.MockResume()
			resume.ResumeID = resumeID
			return resume, nil
		},
	}

	handler := createTestResumeHandler(t, mockRepo)
	handler.CacheDir = t.TempDir()

	var generationCount int32
	handler.GenerateFromResume = func(resume *model.Resume) (string, error) {
		atomic.AddInt32(&generationCount, 1)
		tempDir := t.TempDir()
		return filepath.Join(tempDir, "resume.tex"), nil
	}
	handler.ConvertToPDF = func(filePath string) ([]byte, error) {
		return []byte("generated-pdf"), nil
	}

	req := testutil.CreateRequestWithUserID("GET", "/download/encoded", "test")
	req.SetPathValue("resumeId", "encoded")
	w := httptest.NewRecorder()

	handler.DownloadResumePDF(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if atomic.LoadInt32(&generationCount) != 1 {
		t.Errorf("Expected generation count 1, got %d", atomic.LoadInt32(&generationCount))
	}

	cachedFilePath := handler.resumeCacheFilePath("decoded")
	if _, err := os.Stat(cachedFilePath); err != nil {
		t.Errorf("Expected cached PDF at %s, stat error: %v", cachedFilePath, err)
	}
}

func TestDownloadDefaultResumePDF_UsesCachedResume(t *testing.T) {
	defaultResume := testutil.MockResume()
	defaultResume.ResumeID = "decoded"

	mockRepo := &MockResumeRepository{
		GetDefaultResumeFunc: func(ctx context.Context) (*model.Resume, error) {
			return defaultResume, nil
		},
	}

	handler := createTestResumeHandler(t, mockRepo)
	handler.CacheDir = t.TempDir()

	cachedFilePath := handler.resumeCacheFilePath("decoded")
	if err := os.WriteFile(cachedFilePath, []byte("cached-pdf"), 0644); err != nil {
		t.Fatalf("Failed to write cached PDF: %v", err)
	}

	var generationCount int32
	handler.GenerateFromResume = func(resume *model.Resume) (string, error) {
		atomic.AddInt32(&generationCount, 1)
		tempDir := t.TempDir()
		return filepath.Join(tempDir, "resume.tex"), nil
	}
	handler.ConvertToPDF = func(filePath string) ([]byte, error) {
		return []byte("generated-pdf"), nil
	}

	req := testutil.CreateRequestWithUserID("GET", "/download/default", "")
	w := httptest.NewRecorder()

	handler.DownloadDefaultResumePDF(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if atomic.LoadInt32(&generationCount) != 0 {
		t.Errorf("Expected generation count 0, got %d", atomic.LoadInt32(&generationCount))
	}
	if got := w.Body.String(); got != "cached-pdf" {
		t.Errorf("Expected cached PDF response, got %q", got)
	}
}

func TestResumeCacheFilePath_SanitizesResumeID(t *testing.T) {
	handler := createTestResumeHandler(t, &MockResumeRepository{})
	handler.CacheDir = t.TempDir()

	cacheFilePath := handler.resumeCacheFilePath("../outside/evil")

	if filepath.Dir(cacheFilePath) != handler.CacheDir {
		t.Fatalf("Expected cache file to stay in cache dir, got %s", cacheFilePath)
	}

	base := filepath.Base(cacheFilePath)
	if len(base) != 68 || filepath.Ext(base) != ".pdf" {
		t.Fatalf("Expected hashed pdf filename, got %s", base)
	}
}
