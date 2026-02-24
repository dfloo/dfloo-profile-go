package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/dfloo/dfloo-profile-go/internal/middleware"
	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
	"github.com/dfloo/dfloo-profile-go/internal/scraper"
)

type JobApplicationHandler struct {
	Repo      repository.JobApplicationRepository
	GetUserID func(context.Context) string
	EncodeID  func(string) string
	DecodeID  func(string) (string, error)
}

func NewJobApplicationHandler(repo repository.JobApplicationRepository) *JobApplicationHandler {
	return &JobApplicationHandler{
		Repo:      repo,
		GetUserID: middleware.GetUserID,
		EncodeID:  middleware.EncodeID,
		DecodeID:  middleware.DecodeID,
	}
}

func (h *JobApplicationHandler) GetUserJobApplications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := h.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	jobApplications, err := h.Repo.GetJobApplicationsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Job Applications not found", http.StatusNotFound)
		return
	}

	for i := range jobApplications {
		jobApplications[i].JobApplicationID = h.EncodeID(jobApplications[i].JobApplicationID)
		if jobApplications[i].ResumeID != "" {
			jobApplications[i].ResumeID = h.EncodeID(jobApplications[i].ResumeID)
		}
	}

	json.NewEncoder(w).Encode(jobApplications)
}

func (h *JobApplicationHandler) PostJobApplication(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := h.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	var jobApplication model.JobApplication
	if err := json.NewDecoder(r.Body).Decode(&jobApplication); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if jobApplication.ResumeID != "" {
		decodedResumeID, err := h.DecodeID(jobApplication.ResumeID)
		if err != nil {
			http.Error(w, "Failed to decode resumeID", http.StatusInternalServerError)
			return
		}
		jobApplication.ResumeID = decodedResumeID
	}

	err := h.Repo.CreateJobApplication(r.Context(), &jobApplication, userID)
	if err != nil {
		http.Error(w, "Failed to create Job Application", http.StatusInternalServerError)
		return
	}
	jobApplication.JobApplicationID = h.EncodeID(jobApplication.JobApplicationID)
	if jobApplication.ResumeID != "" {
		jobApplication.ResumeID = h.EncodeID(jobApplication.ResumeID)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&jobApplication)
}

func (h *JobApplicationHandler) PostJobApplicationFromURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := h.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	var request model.FromURLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	parsedURL, err := url.ParseRequestURI(request.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	extract, snapshot, err := scraper.ScrapeJobPosting(r.Context(), request.URL)
	if err != nil {
		http.Error(w, "Failed to scrape job posting", http.StatusBadRequest)
		return
	}

	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		http.Error(w, "Failed to serialize job posting snapshot", http.StatusInternalServerError)
		return
	}

	jobApplication := model.JobApplication{
		Company:     extract.Company,
		Role:        extract.Role,
		Description: extract.Description,
		SourceURL:   request.URL,
		Snapshot:    snapshotBytes,
	}

	err = h.Repo.CreateJobApplication(r.Context(), &jobApplication, userID)
	if err != nil {
		http.Error(w, "Failed to create Job Application", http.StatusInternalServerError)
		return
	}

	jobApplication.JobApplicationID = h.EncodeID(jobApplication.JobApplicationID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&jobApplication)
}

func (h *JobApplicationHandler) PutJobApplications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var jobApplications []*model.JobApplication
	if err := json.NewDecoder(r.Body).Decode(&jobApplications); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := h.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	encodedIDs := make(map[int]struct {
		jobApplicationID string
		resumeID         string
	})

	for i, jobApplication := range jobApplications {
		encodedIDs[i] = struct {
			jobApplicationID string
			resumeID         string
		}{
			jobApplicationID: jobApplication.JobApplicationID,
			resumeID:         jobApplication.ResumeID,
		}

		decodedJobApplicationID, err := h.DecodeID(jobApplication.JobApplicationID)
		if err != nil {
			http.Error(w, "Failed to decode jobApplicationID", http.StatusInternalServerError)
			return
		}
		jobApplication.JobApplicationID = decodedJobApplicationID

		if jobApplication.ResumeID != "" {
			decodedResumeID, err := h.DecodeID(jobApplication.ResumeID)
			if err != nil {
				http.Error(w, "Failed to decode resumeID", http.StatusInternalServerError)
				return
			}
			jobApplication.ResumeID = decodedResumeID
		}
	}

	err := h.Repo.UpdateJobApplications(r.Context(), jobApplications, userID)
	if err != nil {
		http.Error(w, "Failed to update Job Applications", http.StatusInternalServerError)
		return
	}

	for i, jobApplication := range jobApplications {
		jobApplication.JobApplicationID = encodedIDs[i].jobApplicationID
		jobApplication.ResumeID = encodedIDs[i].resumeID
	}

	json.NewEncoder(w).Encode(jobApplications)
}
