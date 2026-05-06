package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dfloo/dfloo-profile-go/internal/claude"
	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/testutil"
)

type MockClaudeClient struct {
	TailorResumeFunc func(ctx context.Context, req claude.TailorRequest) (*claude.TailorResponse, error)
}

func (m *MockClaudeClient) TailorResume(ctx context.Context, req claude.TailorRequest) (*claude.TailorResponse, error) {
	if m.TailorResumeFunc != nil {
		return m.TailorResumeFunc(ctx, req)
	}
	return nil, errors.New("TailorResume not implemented in mock")
}

func createTestTailorHandler(t *testing.T, mockRepo *MockResumeRepository, mockClaude claude.Client) *ResumeHandler {
	t.Helper()
	h := createTestResumeHandler(t, mockRepo)
	h.ClaudeClient = mockClaude
	return h
}

func mockResumeWithExperience() *model.Resume {
	return &model.Resume{
		ResumeID: "resumeId",
		Profile: model.Profile{
			ProfileID: "profileId",
			FirstName: "Devin",
			LastName:  "Flood",
		},
		Summary: "Experienced software engineer",
		Skills:  []string{"Go", "Python", "Docker"},
		Experience: []model.Experience{
			{
				Employer:     "Acme Corp",
				Title:        "Software Engineer",
				StartDate:    "2020-01",
				EndDate:      "2023-01",
				BulletPoints: []string{"Built scalable services", "Led team of 3"},
			},
		},
		Education: []model.Education{
			{Name: "State University", Type: "B.S. Computer Science", CompletionDate: "2019"},
		},
		TemplateSettings: model.TemplateSettings{FontFamily: "lmodern"},
	}
}

func TestTailorResume(t *testing.T) {
	tailoredResponse := &claude.TailorResponse{
		Summary: "Tailored summary for the role",
		Skills:  []string{"Go", "Docker"},
		Experience: []claude.ExperienceOutput{
			{BulletPoints: []string{"Built highly scalable services for fintech"}},
		},
	}

	tests := []struct {
		name         string
		path         string
		body         string
		userID       string
		setupRepo    func(*MockResumeRepository)
		setupClaude  func(*MockClaudeClient)
		wantStatus   int
		checkBody    func(t *testing.T, body []byte)
	}{
		{
			name:   "success",
			path:   "/api/resumes/encoded/tailor",
			body:   `{"jobDescription":"Build fintech services","company":"FinCo","role":"Senior Engineer"}`,
			userID: "user1",
			setupRepo: func(m *MockResumeRepository) {
				m.GetResumeByIDFunc = func(ctx context.Context, resumeID, userID string) (*model.Resume, error) {
					return mockResumeWithExperience(), nil
				}
			},
			setupClaude: func(m *MockClaudeClient) {
				m.TailorResumeFunc = func(ctx context.Context, req claude.TailorRequest) (*claude.TailorResponse, error) {
					return tailoredResponse, nil
				}
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resume model.Resume
				if err := json.Unmarshal(body, &resume); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resume.Summary != tailoredResponse.Summary {
					t.Errorf("summary = %q, want %q", resume.Summary, tailoredResponse.Summary)
				}
				if len(resume.Skills) != len(tailoredResponse.Skills) {
					t.Errorf("skills length = %d, want %d", len(resume.Skills), len(tailoredResponse.Skills))
				}
				if resume.Experience[0].BulletPoints[0] != tailoredResponse.Experience[0].BulletPoints[0] {
					t.Errorf("bullet point = %q, want %q", resume.Experience[0].BulletPoints[0], tailoredResponse.Experience[0].BulletPoints[0])
				}
				// Immutable fields unchanged
				if resume.Experience[0].Employer != "Acme Corp" {
					t.Errorf("employer changed: got %q", resume.Experience[0].Employer)
				}
				if resume.Experience[0].Title != "Software Engineer" {
					t.Errorf("title changed: got %q", resume.Experience[0].Title)
				}
				if len(resume.Education) != 1 || resume.Education[0].Name != "State University" {
					t.Errorf("education changed")
				}
				if resume.TemplateSettings.FontFamily != "lmodern" {
					t.Errorf("template settings changed")
				}
			},
		},
		{
			name:       "no user id",
			path:       "/api/resumes/encoded/tailor",
			body:       `{"jobDescription":"Build fintech services"}`,
			userID:     "",
			setupRepo:  func(m *MockResumeRepository) {},
			setupClaude: func(m *MockClaudeClient) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid json body",
			path:       "/api/resumes/encoded/tailor",
			body:       `not json`,
			userID:     "user1",
			setupRepo:  func(m *MockResumeRepository) {},
			setupClaude: func(m *MockClaudeClient) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing job description",
			path:       "/api/resumes/encoded/tailor",
			body:       `{"company":"FinCo","role":"Engineer"}`,
			userID:     "user1",
			setupRepo:  func(m *MockResumeRepository) {},
			setupClaude: func(m *MockClaudeClient) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "decode id error",
			path:       "/api/resumes/invalid/tailor",
			body:       `{"jobDescription":"Build fintech services"}`,
			userID:     "user1",
			setupRepo:  func(m *MockResumeRepository) {},
			setupClaude: func(m *MockClaudeClient) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "resume not found",
			path:   "/api/resumes/encoded/tailor",
			body:   `{"jobDescription":"Build fintech services"}`,
			userID: "user1",
			setupRepo: func(m *MockResumeRepository) {
				m.GetResumeByIDFunc = func(ctx context.Context, resumeID, userID string) (*model.Resume, error) {
					return nil, errors.New("not found")
				}
			},
			setupClaude: func(m *MockClaudeClient) {},
			wantStatus:  http.StatusNotFound,
		},
		{
			name:   "claude error",
			path:   "/api/resumes/encoded/tailor",
			body:   `{"jobDescription":"Build fintech services"}`,
			userID: "user1",
			setupRepo: func(m *MockResumeRepository) {
				m.GetResumeByIDFunc = func(ctx context.Context, resumeID, userID string) (*model.Resume, error) {
					return mockResumeWithExperience(), nil
				}
			},
			setupClaude: func(m *MockClaudeClient) {
				m.TailorResumeFunc = func(ctx context.Context, req claude.TailorRequest) (*claude.TailorResponse, error) {
					return nil, errors.New("upstream error")
				}
			},
			wantStatus: http.StatusBadGateway,
		},
		{
			name:   "claude experience length mismatch",
			path:   "/api/resumes/encoded/tailor",
			body:   `{"jobDescription":"Build fintech services"}`,
			userID: "user1",
			setupRepo: func(m *MockResumeRepository) {
				m.GetResumeByIDFunc = func(ctx context.Context, resumeID, userID string) (*model.Resume, error) {
					return mockResumeWithExperience(), nil
				}
			},
			setupClaude: func(m *MockClaudeClient) {
				m.TailorResumeFunc = func(ctx context.Context, req claude.TailorRequest) (*claude.TailorResponse, error) {
					return &claude.TailorResponse{
						Summary:    "ok",
						Skills:     []string{"Go"},
						Experience: []claude.ExperienceOutput{}, // wrong length
					}, nil
				}
			},
			wantStatus: http.StatusBadGateway,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &MockResumeRepository{}
			mockClaude := &MockClaudeClient{}
			tc.setupRepo(mockRepo)
			tc.setupClaude(mockClaude)

			h := createTestTailorHandler(t, mockRepo, mockClaude)

			req := testutil.CreateRequestWithBody(http.MethodPost, tc.path, tc.body, tc.userID)
			req.SetPathValue("resumeId", extractResumeIDFromPath(tc.path))

			rr := httptest.NewRecorder()
			h.TailorResume(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tc.wantStatus, rr.Body.String())
			}

			if tc.checkBody != nil {
				tc.checkBody(t, rr.Body.Bytes())
			}
		})
	}
}

// extractResumeIDFromPath extracts the resumeId segment from paths like /api/resumes/{resumeId}/tailor.
func extractResumeIDFromPath(path string) string {
	parts := splitPath(path)
	for i, p := range parts {
		if p == "resumes" && i+2 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func splitPath(path string) []string {
	var parts []string
	for _, p := range splitOnSlash(path) {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitOnSlash(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}
