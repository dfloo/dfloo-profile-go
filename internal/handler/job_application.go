package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dfloo/dfloo-profile-go/internal/middleware"
	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
)

type JobApplicationHandler struct {
	Repo repository.JobApplicationRepository
}

func NewJobApplicationHandler(repo repository.JobApplicationRepository) *JobApplicationHandler {
	return &JobApplicationHandler{Repo: repo}
}

func (h *JobApplicationHandler) GetUserJobApplications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r.Context())
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
		jobApplications[i].JobApplicationID = middleware.EncodeID(jobApplications[i].JobApplicationID)
		if jobApplications[i].ResumeID != "" {
			jobApplications[i].ResumeID = middleware.EncodeID(jobApplications[i].ResumeID)
		}
	}

	json.NewEncoder(w).Encode(jobApplications)
}

func (h *JobApplicationHandler) PostJobApplication(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	var jobApplication model.JobApplication
	if err := json.NewDecoder(r.Body).Decode(&jobApplication); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.Repo.CreateJobApplication(r.Context(), &jobApplication, userID)
	if err != nil {
		http.Error(w, "Failed to create Job Application", http.StatusInternalServerError)
		return
	}
	jobApplication.JobApplicationID = middleware.EncodeID(jobApplication.JobApplicationID)
	jobApplication.ResumeID = middleware.EncodeID(jobApplication.ResumeID)

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

	userID := middleware.GetUserID(r.Context())
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

		decodedJobApplicationID, err := middleware.DecodeID(jobApplication.JobApplicationID)
		if err != nil {
			http.Error(w, "Failed to decode jobApplicationID", http.StatusInternalServerError)
			return
		}
		decodedResumeID, err := middleware.DecodeID(jobApplication.ResumeID)
		if err != nil {
			http.Error(w, "Failed to decode resumeID", http.StatusInternalServerError)
			return
		}
		jobApplication.JobApplicationID = decodedJobApplicationID
		jobApplication.ResumeID = decodedResumeID
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
