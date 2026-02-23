package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dfloo/dfloo-profile-go/internal/latex"
	"github.com/dfloo/dfloo-profile-go/internal/middleware"
	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
)

type ResumeHandler struct {
	Repo          repository.ResumeRepository
	GetUserID     func(context.Context) string
	EncodeID      func(string) string
	DecodeID      func(string) (string, error)
	HasPermission func(context.Context, string) bool
}

func NewResumeHandler(repo repository.ResumeRepository) *ResumeHandler {
	return &ResumeHandler{
		Repo:          repo,
		GetUserID:     middleware.GetUserID,
		EncodeID:      middleware.EncodeID,
		DecodeID:      middleware.DecodeID,
		HasPermission: middleware.HasPermission,
	}
}

func (h *ResumeHandler) GetUserResumes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := h.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	resumes, err := h.Repo.GetResumesByUserID(
		r.Context(),
		userID,
	)
	if err != nil {
		http.Error(w, "Resumes not found", http.StatusNotFound)
		return
	}

	for i := range resumes {
		resumes[i].ResumeID = h.EncodeID(resumes[i].ResumeID)
		resumes[i].Profile.ProfileID = h.EncodeID(resumes[i].Profile.ProfileID)
	}

	json.NewEncoder(w).Encode(resumes)
}

func (h *ResumeHandler) PostResume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := h.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	var resume model.Resume
	if err := json.NewDecoder(r.Body).Decode(&resume); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.Repo.CreateResume(r.Context(), &resume, userID)
	if err != nil {
		http.Error(w, "Failed to create resume", http.StatusInternalServerError)
		return
	}
	resume.ResumeID = h.EncodeID(resume.ResumeID)
	resume.Profile.ProfileID = h.EncodeID(resume.Profile.ProfileID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&resume)
}

func (h *ResumeHandler) PutResume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var resume model.Resume
	if err := json.NewDecoder(r.Body).Decode(&resume); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := h.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	encodedResumeID := resume.ResumeID
	encodedProfileID := resume.Profile.ProfileID
	decodedResumeID, err := h.DecodeID(resume.ResumeID)
	if err != nil {
		http.Error(w, "Failed to decode resumeID", http.StatusInternalServerError)
		return
	}
	decodedProfileID, err := h.DecodeID(resume.Profile.ProfileID)
	if err != nil {
		http.Error(w, "Failed to decode profileID", http.StatusInternalServerError)
		return
	}
	resume.ResumeID = decodedResumeID
	resume.Profile.ProfileID = decodedProfileID

	err = h.Repo.UpdateResume(r.Context(), &resume, userID)
	if err != nil {
		http.Error(w, "Failed to update resume", http.StatusInternalServerError)
		return
	}
	resume.ResumeID = encodedResumeID
	resume.Profile.ProfileID = encodedProfileID

	json.NewEncoder(w).Encode(&resume)
}

func (h *ResumeHandler) SetDefaultResume(w http.ResponseWriter, r *http.Request) {
	userID := h.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	if !h.HasPermission(r.Context(), "set:default_resume") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		ResumeID string `json:"resumeId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	decodedResumeID, err := h.DecodeID(req.ResumeID)
	if err != nil {
		http.Error(w, "Failed to decode resumeID", http.StatusInternalServerError)
		return
	}

	prevDefault, _ := h.Repo.GetDefaultResume(r.Context())

	if prevDefault != nil {
		prevDefault.Default = false
		err = h.Repo.UpdateResume(r.Context(), prevDefault, userID)
		if err != nil {
			http.Error(w, "Failed to update previous default resume", http.StatusInternalServerError)
			return
		}
	}

	newDefault, err := h.Repo.GetResumeByID(r.Context(), decodedResumeID, userID)
	if err != nil {
		http.Error(w, "Resume not found", http.StatusNotFound)
		return
	}
	newDefault.Default = true
	err = h.Repo.UpdateResume(r.Context(), newDefault, userID)
	if err != nil {
		http.Error(w, "Failed to update default resume", http.StatusInternalServerError)
		return
	}
	newDefault.ResumeID = req.ResumeID
	newDefault.Profile.ProfileID = h.EncodeID(newDefault.Profile.ProfileID)

	var result []*model.Resume
	result = append(result, newDefault)
	if prevDefault != nil {
		prevDefault.ResumeID = h.EncodeID(prevDefault.ResumeID)
		prevDefault.Profile.ProfileID = h.EncodeID(prevDefault.Profile.ProfileID)
		result = append(result, prevDefault)
	}
	json.NewEncoder(w).Encode(result)
}

func (h *ResumeHandler) DeleteResumes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var resumeIDs []string
	if err := json.NewDecoder(r.Body).Decode(&resumeIDs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := h.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	for i := range resumeIDs {
		decodedID, err := h.DecodeID(resumeIDs[i])
		if err != nil {
			http.Error(w, "Failed to decode resumeID", http.StatusInternalServerError)
			return
		}
		resumeIDs[i] = decodedID
	}

	deletedIDs, err := h.Repo.DeleteResumes(r.Context(), resumeIDs, userID)
	if err != nil {
		http.Error(w, "Failed to delete resumes", http.StatusInternalServerError)
		return
	}

	for i := range deletedIDs {
		deletedIDs[i] = h.EncodeID(deletedIDs[i])
	}

	json.NewEncoder(w).Encode(deletedIDs)
}

func (h *ResumeHandler) DownloadDefaultResumePDF(w http.ResponseWriter, r *http.Request) {
	resume, err := h.Repo.GetDefaultResume(r.Context())
	if err != nil {
		http.Error(w, "Default resume not found", http.StatusNotFound)
		return
	}

	DownloadResume(w, resume)
}

func (h *ResumeHandler) DownloadResumePDF(w http.ResponseWriter, r *http.Request) {
	var resume model.Resume
	if err := json.NewDecoder(r.Body).Decode(&resume); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	DownloadResume(w, &resume)
}

func DownloadResume(w http.ResponseWriter, resume *model.Resume) {
	hash, err := ResumeHash(resume)
	if err != nil {
		http.Error(w, "Failed to generate resume hash", http.StatusInternalServerError)
		return
	}

	cachePath := "/tmp/resume_cache/"
	cachedHashPath := cachePath + hash + ".pdf"
	var pdfBytes []byte
	if _, err := os.Stat(cachedHashPath); err == nil {
		log.Print("Serving cached resume pdf")
		pdfBytes, err = os.ReadFile(cachedHashPath)
		if err != nil {
			log.Printf("failed to read cached resume pdf: %v", err)
			http.Error(w, "Failed to read resume PDF cache", http.StatusInternalServerError)
			return
		}
	} else {
		log.Print("Generating resume pdf")
		filePath, err := latex.GenerateFromResume(resume)
		if err != nil {
			log.Printf("failed to generate resume tex file: %v", err)
			http.Error(w, "Failed to prepare resume PDF", http.StatusInternalServerError)
			return
		}

		tempDir := filepath.Dir(filePath)
		defer func() {
			if tempDir != "" {
				if removeErr := os.RemoveAll(tempDir); removeErr != nil {
					log.Printf("failed to remove temporary resume dir: %v", removeErr)
				}
			}
		}()

		pdfBytes, err = latex.ConvertToPDF(filePath)
		if err != nil {
			log.Printf("failed to convert resume tex to pdf: %v", err)
			http.Error(w, "Failed to generate resume PDF", http.StatusInternalServerError)
			return
		}
		if err := os.MkdirAll(cachePath, 0755); err != nil {
			log.Printf("failed to create resume cache dir: %v", err)
		} else if err := os.WriteFile(cachedHashPath, pdfBytes, 0644); err != nil {
			log.Printf("failed to write resume cache file: %v", err)
		}
	}
	if err := CleanUpOldCacheFiles(cachePath, 24*time.Hour); err != nil {
		log.Printf("failed to cleanup old resume cache files: %v", err)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=resume.pdf")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(pdfBytes); err != nil {
		log.Printf("failed to write resume pdf response: %v", err)
	}
}

func ResumeHash(resume *model.Resume) (string, error) {
	data, err := json.Marshal(resume)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func CleanUpOldCacheFiles(cacheDir string, maxAge time.Duration) error {
	now := time.Now()
	return filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && now.Sub(info.ModTime()) > maxAge {
			return os.Remove(path)
		}
		return nil
	})
}
