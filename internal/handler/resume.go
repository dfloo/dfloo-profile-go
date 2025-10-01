package handler

import (
	"encoding/json"
	"log"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
)

type ResumeHandler struct {
	Repo *repository.ResumeRepository
}

func NewResumeHandler(repo *repository.ResumeRepository) *ResumeHandler {
	return &ResumeHandler{Repo: repo}
}

func (h *ResumeHandler) GetUserResumes(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(
		jwtmiddleware.ContextKey{},
	).(*validator.ValidatedClaims)
	if !ok || claims == nil || claims.RegisteredClaims.Subject == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	resumes, err := h.Repo.GetResumesByUserID(
		r.Context(),
		claims.RegisteredClaims.Subject,
	)
	if err != nil {
		http.Error(w, "Resumes not found", http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resumes)
}

func (h *ResumeHandler) PostResume(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok || claims == nil || claims.RegisteredClaims.Subject == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	userID := claims.RegisteredClaims.Subject
	var resume model.Resume
	if err := json.NewDecoder(r.Body).Decode(&resume); err != nil {
		log.Printf("post resume: %v %v", resume, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.Repo.CreateResume(r.Context(), &resume, userID)
	if err != nil {
		http.Error(w, "Failed to create resume", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&resume)
}

func (h *ResumeHandler) PutResume(w http.ResponseWriter, r *http.Request) {
	var resume model.Resume
	if err := json.NewDecoder(r.Body).Decode(&resume); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.Repo.UpdateResume(r.Context(), &resume)
	if err != nil {
		http.Error(w, "Failed to update resume", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&resume)
}

func (h *ResumeHandler) DeleteResumes(w http.ResponseWriter, r *http.Request) {
	var resumeIDs []string
	if err := json.NewDecoder(r.Body).Decode(&resumeIDs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Repo.DeleteResumes(r.Context(), resumeIDs); err != nil {
		http.Error(w, "Failed to delete resumes", http.StatusInternalServerError)
		return
	}
}
