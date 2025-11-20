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

func createTestJobApplicationHandler(mockRepo *MockJobApplicationRepository) *JobApplicationHandler {
	return &JobApplicationHandler{
		Repo:      mockRepo,
		GetUserID: testutil.MockGetUserID,
		EncodeID:  testutil.MockEncodeID,
		DecodeID:  testutil.MockDecodeID,
	}
}

type MockJobApplicationRepository struct {
	GetJobApplicationsByUserIDFunc func(ctx context.Context, userID string) ([]*model.JobApplication, error)
	CreateJobApplicationFunc       func(ctx context.Context, jobApplication *model.JobApplication, userID string) error
	UpdateJobApplicationsFunc      func(ctx context.Context, jobApplications []*model.JobApplication, userID string) error
}

func (m *MockJobApplicationRepository) GetJobApplicationsByUserID(ctx context.Context, userID string) ([]*model.JobApplication, error) {
	if m.GetJobApplicationsByUserIDFunc != nil {
		return m.GetJobApplicationsByUserIDFunc(ctx, userID)
	}
	return []*model.JobApplication{}, errors.New("GetJobApplicationsByUserID not implemented in mock")
}

func (m *MockJobApplicationRepository) CreateJobApplication(ctx context.Context, jobApplication *model.JobApplication, userID string) error {
	if m.CreateJobApplicationFunc != nil {
		return m.CreateJobApplicationFunc(ctx, jobApplication, userID)
	}
	return errors.New("CreateJobApplication not implemented in mock")
}

func (m *MockJobApplicationRepository) UpdateJobApplications(ctx context.Context, jobApplications []*model.JobApplication, userID string) error {
	if m.UpdateJobApplicationsFunc != nil {
		return m.UpdateJobApplicationsFunc(ctx, jobApplications, userID)
	}
	return errors.New("UpdateJobApplications not implemented in mock")
}

func TestGetUserJobApplications_Success(t *testing.T) {
	mockRepo := &MockJobApplicationRepository{
		GetJobApplicationsByUserIDFunc: func(ctx context.Context, userID string) ([]*model.JobApplication, error) {
			return []*model.JobApplication{testutil.MockJobApplication()}, nil
		},
	}

	handler := createTestJobApplicationHandler(mockRepo)
	req := testutil.CreateRequestWithUserID("GET", "/job-applications", "test")
	w := httptest.NewRecorder()

	handler.GetUserJobApplications(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var jobApplications []*model.JobApplication
	err := json.NewDecoder(w.Body).Decode(&jobApplications)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(jobApplications) != 1 {
		t.Errorf("Expected 1 job application, got %d", len(jobApplications))
	}

	if jobApplications[0].JobApplicationID != "encoded" {
		t.Errorf("Expected JobApplicationID 'encoded', got '%s'", jobApplications[0].JobApplicationID)
	}

	if jobApplications[0].ResumeID != "encoded" {
		t.Errorf("Expected ResumeID 'encoded', got '%s'", jobApplications[0].ResumeID)
	}
}

func TestGetUserJobApplications_SuccessNoResumeID(t *testing.T) {
	mockRepo := &MockJobApplicationRepository{
		GetJobApplicationsByUserIDFunc: func(ctx context.Context, userID string) ([]*model.JobApplication, error) {
			jobApp := testutil.MockJobApplication()
			jobApp.ResumeID = ""
			return []*model.JobApplication{jobApp}, nil
		},
	}

	handler := createTestJobApplicationHandler(mockRepo)
	req := testutil.CreateRequestWithUserID("GET", "/job-applications", "test")
	w := httptest.NewRecorder()

	handler.GetUserJobApplications(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var jobApplications []*model.JobApplication
	err := json.NewDecoder(w.Body).Decode(&jobApplications)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if jobApplications[0].ResumeID != "" {
		t.Errorf("Expected empty ResumeID, got '%s'", jobApplications[0].ResumeID)
	}
}

func TestGetUserJobApplications_NoUserID(t *testing.T) {
	handler := createTestJobApplicationHandler(&MockJobApplicationRepository{})
	req := testutil.CreateRequestWithUserID("GET", "/job-applications", "")
	w := httptest.NewRecorder()

	handler.GetUserJobApplications(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetUserJobApplications_NotFound(t *testing.T) {
	mockRepo := &MockJobApplicationRepository{
		GetJobApplicationsByUserIDFunc: func(ctx context.Context, userID string) ([]*model.JobApplication, error) {
			return nil, errors.New("no job applications found")
		},
	}

	handler := createTestJobApplicationHandler(mockRepo)
	req := testutil.CreateRequestWithUserID("GET", "/job-applications", "test")
	w := httptest.NewRecorder()

	handler.GetUserJobApplications(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestPostJobApplication_Success(t *testing.T) {
	mockRepo := &MockJobApplicationRepository{
		CreateJobApplicationFunc: func(ctx context.Context, jobApplication *model.JobApplication, userID string) error {
			jobApplication.JobApplicationID = "newJobApplicationId"
			return nil
		},
	}

	handler := createTestJobApplicationHandler(mockRepo)
	jobAppJSON := testutil.MockJobApplicationJSON()
	req := testutil.CreateRequestWithBody("POST", "/job-applications", jobAppJSON, "test")
	w := httptest.NewRecorder()

	handler.PostJobApplication(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var jobApplication model.JobApplication
	err := json.NewDecoder(w.Body).Decode(&jobApplication)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if jobApplication.JobApplicationID != "encoded" {
		t.Errorf("Expected JobApplicationID 'encoded', got '%s'", jobApplication.JobApplicationID)
	}

	if jobApplication.ResumeID != "encoded" {
		t.Errorf("Expected ResumeID 'encoded', got '%s'", jobApplication.ResumeID)
	}
}

func TestPostJobApplication_SuccessWithEmptyResumeID(t *testing.T) {
	mockRepo := &MockJobApplicationRepository{
		CreateJobApplicationFunc: func(ctx context.Context, jobApplication *model.JobApplication, userID string) error {
			jobApplication.JobApplicationID = "newJobApplicationId"
			return nil
		},
	}

	handler := createTestJobApplicationHandler(mockRepo)
	jobAppJSON := testutil.MockJobApplicationWithoutResumeJSON()
	req := testutil.CreateRequestWithBody("POST", "/job-applications", jobAppJSON, "test")
	w := httptest.NewRecorder()

	handler.PostJobApplication(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var jobApplication model.JobApplication
	err := json.NewDecoder(w.Body).Decode(&jobApplication)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if jobApplication.ResumeID != "" {
		t.Errorf("Expected empty ResumeID, got '%s'", jobApplication.ResumeID)
	}
}

func TestPostJobApplication_NoUserID(t *testing.T) {
	handler := createTestJobApplicationHandler(&MockJobApplicationRepository{})
	req := testutil.CreateRequestWithBody("POST", "/job-applications", "{}", "")
	w := httptest.NewRecorder()

	handler.PostJobApplication(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPostJobApplication_InvalidJSON(t *testing.T) {
	handler := createTestJobApplicationHandler(&MockJobApplicationRepository{})
	req := testutil.CreateRequestWithBody("POST", "/job-applications", "invalid json", "test")
	w := httptest.NewRecorder()

	handler.PostJobApplication(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPostJobApplication_RepositoryError(t *testing.T) {
	mockRepo := &MockJobApplicationRepository{
		CreateJobApplicationFunc: func(ctx context.Context, jobApplication *model.JobApplication, userID string) error {
			return errors.New("database error")
		},
	}

	handler := createTestJobApplicationHandler(mockRepo)
	req := testutil.CreateRequestWithBody("POST", "/job-applications", "{}", "test")
	w := httptest.NewRecorder()

	handler.PostJobApplication(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPutJobApplications_Success(t *testing.T) {
	mockRepo := &MockJobApplicationRepository{
		UpdateJobApplicationsFunc: func(ctx context.Context, jobApplications []*model.JobApplication, userID string) error {
			for _, jobApp := range jobApplications {
				if jobApp.JobApplicationID != "decoded" {
					return errors.New("JobApplicationID was not decoded")
				}
				if jobApp.ResumeID != "" && jobApp.ResumeID != "decoded" {
					return errors.New("ResumeID was not decoded")
				}
			}
			return nil
		},
	}

	handler := createTestJobApplicationHandler(mockRepo)
	jobAppsJSON := testutil.MockJobApplicationsArrayJSON()
	req := testutil.CreateRequestWithBody("PUT", "/job-applications", jobAppsJSON, "test")
	w := httptest.NewRecorder()

	handler.PutJobApplications(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var jobApplications []*model.JobApplication
	err := json.NewDecoder(w.Body).Decode(&jobApplications)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	for _, jobApp := range jobApplications {
		if jobApp.JobApplicationID != "encoded" {
			t.Errorf("Expected JobApplicationID 'encoded', got '%s'", jobApp.JobApplicationID)
		}
	}
}

func TestPutJobApplications_SuccessWithMixedResumeIDs(t *testing.T) {
	mockRepo := &MockJobApplicationRepository{
		UpdateJobApplicationsFunc: func(ctx context.Context, jobApplications []*model.JobApplication, userID string) error {
			return nil
		},
	}

	handler := createTestJobApplicationHandler(mockRepo)

	jobAppsJSON := `[
        {
            "jobApplicationId": "job1",
            "resumeId": "resume1",
            "company": "Company A",
            "role": "Developer"
        },
        {
            "jobApplicationId": "job2",
            "resumeId": "",
            "company": "Company B",
            "role": "Engineer"
        }
    ]`
	req := testutil.CreateRequestWithBody("PUT", "/job-applications", jobAppsJSON, "test")
	w := httptest.NewRecorder()

	handler.PutJobApplications(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestPutJobApplications_InvalidJSON(t *testing.T) {
	handler := createTestJobApplicationHandler(&MockJobApplicationRepository{})
	req := testutil.CreateRequestWithBody("PUT", "/job-applications", "invalid json", "test")
	w := httptest.NewRecorder()

	handler.PutJobApplications(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPutJobApplications_NoUserID(t *testing.T) {
	handler := createTestJobApplicationHandler(&MockJobApplicationRepository{})
	req := testutil.CreateRequestWithBody("PUT", "/job-applications", "[]", "")
	w := httptest.NewRecorder()

	handler.PutJobApplications(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPutJobApplications_DecodeJobApplicationIDError(t *testing.T) {
	handler := createTestJobApplicationHandler(&MockJobApplicationRepository{})

	jobAppsJSON := `[{"jobApplicationId": "invalid", "resumeId": "resume1"}]`
	req := testutil.CreateRequestWithBody("PUT", "/job-applications", jobAppsJSON, "test")
	w := httptest.NewRecorder()

	handler.PutJobApplications(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPutJobApplications_DecodeResumeIDError(t *testing.T) {
	handler := createTestJobApplicationHandler(&MockJobApplicationRepository{})

	jobAppsJSON := `[{"jobApplicationId": "job1", "resumeId": "invalid"}]`
	req := testutil.CreateRequestWithBody("PUT", "/job-applications", jobAppsJSON, "test")
	w := httptest.NewRecorder()

	handler.PutJobApplications(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPutJobApplications_RepositoryError(t *testing.T) {
	mockRepo := &MockJobApplicationRepository{
		UpdateJobApplicationsFunc: func(ctx context.Context, jobApplications []*model.JobApplication, userID string) error {
			return errors.New("update failed")
		},
	}

	handler := createTestJobApplicationHandler(mockRepo)
	jobAppsJSON := testutil.MockJobApplicationsArrayJSON()
	req := testutil.CreateRequestWithBody("PUT", "/job-applications", jobAppsJSON, "test")
	w := httptest.NewRecorder()

	handler.PutJobApplications(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPutJobApplications_EmptyArray(t *testing.T) {
	mockRepo := &MockJobApplicationRepository{
		UpdateJobApplicationsFunc: func(ctx context.Context, jobApplications []*model.JobApplication, userID string) error {
			if len(jobApplications) != 0 {
				return errors.New("expected empty array")
			}
			return nil
		},
	}

	handler := createTestJobApplicationHandler(mockRepo)
	req := testutil.CreateRequestWithBody("PUT", "/job-applications", "[]", "test")
	w := httptest.NewRecorder()

	handler.PutJobApplications(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetUserJobApplications_EmptyResult(t *testing.T) {
	mockRepo := &MockJobApplicationRepository{
		GetJobApplicationsByUserIDFunc: func(ctx context.Context, userID string) ([]*model.JobApplication, error) {
			return []*model.JobApplication{}, nil
		},
	}

	handler := createTestJobApplicationHandler(mockRepo)
	req := testutil.CreateRequestWithUserID("GET", "/job-applications", "test")
	w := httptest.NewRecorder()

	handler.GetUserJobApplications(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var jobApplications []*model.JobApplication
	err := json.NewDecoder(w.Body).Decode(&jobApplications)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(jobApplications) != 0 {
		t.Errorf("Expected 0 job applications, got %d", len(jobApplications))
	}
}
