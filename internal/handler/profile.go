package handler

import (
	"encoding/json"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
)

type ProfileHandler struct {
	Repo *repository.ProfileRepository
}

func NewProfileHandler(repo *repository.ProfileRepository) *ProfileHandler {
	return &ProfileHandler{Repo: repo}
}

func (h *ProfileHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok || claims == nil || claims.RegisteredClaims.Subject == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	profile, err := h.Repo.GetProfileByUserID(r.Context(), claims.RegisteredClaims.Subject)
	if err != nil {
		profile = &model.Profile{}
	}

	json.NewEncoder(w).Encode(profile)
}

func (h *ProfileHandler) PostUserProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok || claims == nil || claims.RegisteredClaims.Subject == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	userID := claims.RegisteredClaims.Subject
	var profile model.Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.Repo.CreateUserProfile(r.Context(), &profile, userID)
	if err != nil {
		http.Error(w, "Failed to create profile", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&profile)
}

func (h *ProfileHandler) PutUserProfile(w http.ResponseWriter, r *http.Request) {
	var profile model.Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.Repo.UpdateProfile(r.Context(), &profile)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&profile)
}
