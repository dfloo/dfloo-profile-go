package testutil

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/dfloo/dfloo-profile-go/internal/model"
)

type ContextKey string

const UserIDKey ContextKey = "userID"

func MockGetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

func MockEncodeID(s string) string {
	return "encoded"
}

func MockDecodeID(s string) (string, error) {
	if s == "invalid" {
		return "", errors.New("decode error")
	}
	return "decoded", nil
}

func CreateRequestWithUserID(method, url, userID string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	if userID != "" {
		ctx := context.WithValue(req.Context(), UserIDKey, userID)
		req = req.WithContext(ctx)
	}
	return req
}

func CreateRequestWithBody(method, url, body, userID string) *http.Request {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		ctx := context.WithValue(req.Context(), UserIDKey, userID)
		req = req.WithContext(ctx)
	}
	return req
}

func MockProfile() *model.Profile {
	return &model.Profile{
		ProfileID: "profileId",
		FirstName: "Devin",
		LastName:  "Flood",
	}
}

func MockProfileJSON() string {
	return `{
		"profileId": "profileId",
		"firstName": "Devin",
		"lastName": "Flood"
	}`
}

func MockResume() *model.Resume {
	return &model.Resume{
		ResumeID:    "resumeId",
		Description: "Mock Resume",
		Profile:     *MockProfile(),
	}
}

func MockResumeJSON() string {
	return `{
		"resumeId": "resumeId",
		"description": "Mock Resume",
		"profile": {
			"profileId": "profileId",
			"firstName": "Devin",
			"lastName": "Flood"
		}
	}`
}

func MockJobApplication() *model.JobApplication {
	return &model.JobApplication{
		JobApplicationID: "jobApplicationId",
		ResumeID:         "resumeId",
		Status:           "Submitted",
		Company:          "Tech Corp",
		Role:             "Sr. Software Engineer",
		Description:      "Great opportunity",
		Notes:            "Interesting company",
	}
}

func MockJobApplicationJSON() string {
	return `{
        "jobApplicationId": "jobApplicationId",
        "resumeId": "resumeId",
        "status": "Submitted",
        "company": "Tech Corp",
        "role": "Sr. Software Engineer",
        "description": "Great opportunity",
        "notes": "Interesting company"
    }`
}

func MockJobApplicationWithoutResumeJSON() string {
	return `{
        "jobApplicationId": "jobApplicationId",
        "resumeId": "",
        "status": "Submitted",
        "company": "Tech Corp",
        "role": "Sr. Software Engineer",
        "description": "Great opportunity",
        "notes": "Interesting company"
    }`
}

func MockJobApplicationsArrayJSON() string {
	return `[
        {
            "jobApplicationId": "encoded",
            "resumeId": "resume1",
            "status": "Submitted",
            "company": "Tech Corp",
            "role": "Sr. Software Engineer"
        },
        {
            "jobApplicationId": "encoded",
            "resumeId": "resume2", 
            "status": "Interview",
            "company": "StartupCo",
            "role": "Full Stack Developer"
        }
    ]`
}
