package handler

import (
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
	Repo *repository.ResumeRepository
}

func NewResumeHandler(repo *repository.ResumeRepository) *ResumeHandler {
	return &ResumeHandler{Repo: repo}
}

func (h *ResumeHandler) GetUserResumes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r.Context())
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
		resumes[i].ResumeID = middleware.EncodeID(resumes[i].ResumeID)
	}

	json.NewEncoder(w).Encode(resumes)
}

func (h *ResumeHandler) PostResume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r.Context())
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
	resume.ResumeID = middleware.EncodeID(resume.ResumeID)

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

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	encoded := resume.ResumeID
	decoded, err := middleware.DecodeID(resume.ResumeID)
	if err != nil {
		http.Error(w, "Failed to decode resumeID", http.StatusInternalServerError)
		return
	}
	resume.ResumeID = decoded

	err = h.Repo.UpdateResume(r.Context(), &resume, userID)
	if err != nil {
		http.Error(w, "Failed to update resume", http.StatusInternalServerError)
		return
	}
	resume.ResumeID = encoded

	json.NewEncoder(w).Encode(&resume)
}

func (h *ResumeHandler) DeleteResumes(w http.ResponseWriter, r *http.Request) {
	var resumeIDs []string
	if err := json.NewDecoder(r.Body).Decode(&resumeIDs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	for i := range resumeIDs {
		decoded, err := middleware.DecodeID(resumeIDs[i])
		if err != nil {
			http.Error(w, "Failed to decode resumeID", http.StatusInternalServerError)
			return
		}
		resumeIDs[i] = string(decoded)
	}

	if err := h.Repo.DeleteResumes(r.Context(), resumeIDs, userID); err != nil {
		http.Error(w, "Failed to delete resumes", http.StatusInternalServerError)
		return
	}
}

func (h *ResumeHandler) DownloadResumePDF(w http.ResponseWriter, r *http.Request) {
	var resume model.Resume
	if err := json.NewDecoder(r.Body).Decode(&resume); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hash, _ := ResumeHash(&resume)
	cachedPath := "/tmp/resume_cache/" + hash + ".pdf"
	var pdfBytes []byte
	if _, err := os.Stat(cachedPath); err == nil {
		log.Print("Serving cached resume pdf")
		pdfBytes, _ = os.ReadFile(cachedPath)
	} else {
		log.Print("Generating resume pdf")
		filePath, _ := latex.GenerateFromResume(&resume)
		pdfBytes, _ = latex.ConvertToPDF(filePath)
		os.MkdirAll("/tmp/resume_cache", 0755)
		os.WriteFile(cachedPath, pdfBytes, 0644)
		defer os.RemoveAll(filePath[:len(filePath)-len("resume.tex")])
	}
	CleanUpOldCacheFiles("/tmp/resume_cache", 24*time.Hour)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=resume.pdf")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
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
