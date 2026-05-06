package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/dfloo/dfloo-profile-go/internal/claude"
	"github.com/dfloo/dfloo-profile-go/internal/latex"
	"github.com/dfloo/dfloo-profile-go/internal/middleware"
	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
)

type ResumeHandler struct {
	Repo          repository.ResumeRepository
	ClaudeClient  claude.Client
	GetUserID     func(context.Context) string
	EncodeID      func(string) string
	DecodeID      func(string) (string, error)
	HasPermission func(context.Context, string) bool
	CacheDir      string

	GenerateFromResume func(*model.Resume) (string, error)
	ConvertToPDF       func(string) ([]byte, error)

	inflightMu  sync.Mutex
	inflightPDF map[string]*resumePDFGeneration
	asyncWg     sync.WaitGroup
}

type resumePDFGeneration struct {
	done chan struct{}
	err  error
}

func NewResumeHandler(repo repository.ResumeRepository) *ResumeHandler {
	return &ResumeHandler{
		Repo:          repo,
		GetUserID:     middleware.GetUserID,
		EncodeID:      middleware.EncodeID,
		DecodeID:      middleware.DecodeID,
		HasPermission: middleware.HasPermission,
		CacheDir:      getResumeCacheDir(),

		GenerateFromResume: latex.GenerateFromResume,
		ConvertToPDF:       latex.ConvertToPDF,
		inflightPDF:        make(map[string]*resumePDFGeneration),
	}
}

func getResumeCacheDir() string {
	if cacheDir := os.Getenv("RESUME_CACHE_PATH"); cacheDir != "" {
		return cacheDir
	}
	return "/tmp/resume_cache"
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
	h.triggerAsyncResumePDFGeneration(&resume)
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
	h.deleteResumeCacheFiles([]string{resume.ResumeID})
	h.triggerAsyncResumePDFGeneration(&resume)
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
		h.deleteResumeCacheFiles([]string{prevDefault.ResumeID})
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
	h.deleteResumeCacheFiles([]string{newDefault.ResumeID})
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

	h.deleteResumeCacheFiles(deletedIDs)

	for i := range deletedIDs {
		deletedIDs[i] = h.EncodeID(deletedIDs[i])
	}

	json.NewEncoder(w).Encode(deletedIDs)
}

func (h *ResumeHandler) DownloadGuestResumePDF(w http.ResponseWriter, r *http.Request) {
	var resume model.Resume
	if err := json.NewDecoder(r.Body).Decode(&resume); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	generateFromResume := h.GenerateFromResume
	if generateFromResume == nil {
		generateFromResume = latex.GenerateFromResume
	}
	convertToPDF := h.ConvertToPDF
	if convertToPDF == nil {
		convertToPDF = latex.ConvertToPDF
	}

	filePath, err := generateFromResume(&resume)
	if err != nil {
		log.Printf("failed to generate latex for guest resume: %v", err)
		http.Error(w, "Failed to generate resume PDF", http.StatusInternalServerError)
		return
	}
	defer func() {
		if tempDir := filepath.Dir(filePath); tempDir != "" {
			if removeErr := os.RemoveAll(tempDir); removeErr != nil {
				log.Printf("failed to remove temporary resume dir: %v", removeErr)
			}
		}
	}()

	pdfBytes, err := convertToPDF(filePath)
	if err != nil {
		log.Printf("failed to convert latex to pdf for guest resume: %v", err)
		http.Error(w, "Failed to generate resume PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=resume.pdf")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(pdfBytes); err != nil {
		log.Printf("failed to write guest resume pdf response: %v", err)
	}
}

func (h *ResumeHandler) DownloadDefaultResumePDF(w http.ResponseWriter, r *http.Request) {
	resume, err := h.Repo.GetDefaultResume(r.Context())
	if err != nil {
		http.Error(w, "Default resume not found", http.StatusNotFound)
		return
	}

	h.downloadResume(w, resume)
}

func (h *ResumeHandler) DownloadResumePDF(w http.ResponseWriter, r *http.Request) {
	encodedResumeID := r.PathValue("resumeId")
	if encodedResumeID == "" {
		http.Error(w, "resumeId is required", http.StatusBadRequest)
		return
	}

	userID := h.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	decodedResumeID, err := h.DecodeID(encodedResumeID)
	if err != nil {
		http.Error(w, "Failed to decode resumeID", http.StatusBadRequest)
		return
	}

	resume, err := h.Repo.GetResumeByID(r.Context(), decodedResumeID, userID)
	if err != nil {
		http.Error(w, "Resume not found", http.StatusNotFound)
		return
	}

	h.downloadResume(w, resume)
}

func (h *ResumeHandler) downloadResume(w http.ResponseWriter, resume *model.Resume) {
	if resume == nil || resume.ResumeID == "" {
		http.Error(w, "resumeId is required", http.StatusBadRequest)
		return
	}

	pdfBytes, err := h.getOrGenerateResumePDF(resume)
	if err != nil {
		log.Printf("failed to get or generate resume pdf: %v", err)
		http.Error(w, "Failed to generate resume PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=resume.pdf")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(pdfBytes); err != nil {
		log.Printf("failed to write resume pdf response: %v", err)
	}
}

func (h *ResumeHandler) triggerAsyncResumePDFGeneration(resume *model.Resume) {
	if resume == nil || resume.ResumeID == "" {
		return
	}

	resumeCopy := *resume
	h.asyncWg.Add(1)
	go func() {
		defer h.asyncWg.Done()
		if err := h.generateResumePDFWithInflight(&resumeCopy); err != nil {
			log.Printf("failed async resume pdf generation for resumeID=%s: %v", resumeCopy.ResumeID, err)
		}
	}()
}

func (h *ResumeHandler) getOrGenerateResumePDF(resume *model.Resume) ([]byte, error) {
	inflightKey, cacheFilePath := h.resumeCacheIdentity(resume)
	if inflightKey == "" || cacheFilePath == "" {
		return nil, errors.New("resume cache identity is invalid")
	}

	pdfBytes, err := os.ReadFile(cacheFilePath)
	if err == nil {
		return pdfBytes, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	generation, owner := h.registerInflightGeneration(inflightKey)
	if owner {
		generation.err = h.generateResumePDF(resume)
		h.completeInflightGeneration(inflightKey)
	} else {
		<-generation.done
	}

	if generation.err != nil {
		return nil, generation.err
	}

	return os.ReadFile(cacheFilePath)
}

func (h *ResumeHandler) generateResumePDFWithInflight(resume *model.Resume) error {
	inflightKey, cacheFilePath := h.resumeCacheIdentity(resume)
	if inflightKey == "" || cacheFilePath == "" {
		return errors.New("resume cache identity is invalid")
	}

	if _, err := os.Stat(cacheFilePath); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	generation, owner := h.registerInflightGeneration(inflightKey)
	if owner {
		generation.err = h.generateResumePDF(resume)
		h.completeInflightGeneration(inflightKey)
		return generation.err
	}

	<-generation.done
	if generation.err != nil {
		return generation.err
	}

	_, err := os.Stat(cacheFilePath)
	if err == nil {
		return nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return errors.New("resume pdf generation completed without cache file")
	}
	return err
}

func (h *ResumeHandler) GetFontOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latex.SupportedFonts())
}

func (h *ResumeHandler) generateResumePDF(resume *model.Resume) error {
	cacheDir := h.getCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	generateFromResume := h.GenerateFromResume
	if generateFromResume == nil {
		generateFromResume = latex.GenerateFromResume
	}
	convertToPDF := h.ConvertToPDF
	if convertToPDF == nil {
		convertToPDF = latex.ConvertToPDF
	}

	filePath, err := generateFromResume(resume)
	if err != nil {
		return err
	}
	tempDir := filepath.Dir(filePath)
	defer func() {
		if tempDir != "" {
			if removeErr := os.RemoveAll(tempDir); removeErr != nil {
				log.Printf("failed to remove temporary resume dir: %v", removeErr)
			}
		}
	}()

	pdfBytes, err := convertToPDF(filePath)
	if err != nil {
		return err
	}

	_, cacheFilePath := h.resumeCacheIdentity(resume)
	if cacheFilePath == "" {
		return errors.New("resume cache identity is invalid")
	}
	tempCacheFilePath := cacheFilePath + ".tmp"
	if err := os.WriteFile(tempCacheFilePath, pdfBytes, 0644); err != nil {
		return err
	}
	if err := os.Rename(tempCacheFilePath, cacheFilePath); err != nil {
		_ = os.Remove(tempCacheFilePath)
		return err
	}

	return nil
}

func (h *ResumeHandler) registerInflightGeneration(inflightKey string) (*resumePDFGeneration, bool) {
	h.inflightMu.Lock()
	defer h.inflightMu.Unlock()

	if h.inflightPDF == nil {
		h.inflightPDF = make(map[string]*resumePDFGeneration)
	}

	if generation, exists := h.inflightPDF[inflightKey]; exists {
		return generation, false
	}

	generation := &resumePDFGeneration{done: make(chan struct{})}
	h.inflightPDF[inflightKey] = generation
	return generation, true
}

func (h *ResumeHandler) completeInflightGeneration(inflightKey string) {
	h.inflightMu.Lock()
	generation, exists := h.inflightPDF[inflightKey]
	if exists {
		delete(h.inflightPDF, inflightKey)
	}
	h.inflightMu.Unlock()

	if exists {
		close(generation.done)
	}
}

func (h *ResumeHandler) resumeCacheFilePath(resumeID string) string {
	return filepath.Join(h.getCacheDir(), safeResumeCacheFilename(resumeID))
}

func (h *ResumeHandler) resumeCacheIdentity(resume *model.Resume) (string, string) {
	if resume == nil || resume.ResumeID == "" {
		return "", ""
	}

	if resume.Updated.IsZero() {
		return resume.ResumeID, h.resumeCacheFilePath(resume.ResumeID)
	}

	versionToken := strconv.FormatInt(resume.Updated.UTC().UnixNano(), 10)
	cacheFilename := safeVersionedResumeCacheFilename(resume.ResumeID, versionToken)
	inflightKey := resume.ResumeID + "|" + versionToken
	return inflightKey, filepath.Join(h.getCacheDir(), cacheFilename)
}

func (h *ResumeHandler) deleteResumeCacheFiles(resumeIDs []string) {
	for _, resumeID := range resumeIDs {
		legacyCacheFilePath := h.resumeCacheFilePath(resumeID)
		if err := os.Remove(legacyCacheFilePath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			log.Printf("failed to remove legacy resume cache file for resumeID=%s: %v", resumeID, err)
		}

		versionedPattern := filepath.Join(h.getCacheDir(), safeResumeIDPrefix(resumeID)+"_*.pdf")
		versionedCacheFilePaths, err := filepath.Glob(versionedPattern)
		if err != nil {
			log.Printf("failed to list versioned resume cache files for resumeID=%s: %v", resumeID, err)
			continue
		}

		for _, cacheFilePath := range versionedCacheFilePaths {
			if err := os.Remove(cacheFilePath); err != nil && !errors.Is(err, fs.ErrNotExist) {
				log.Printf("failed to remove versioned resume cache file for resumeID=%s path=%s: %v", resumeID, cacheFilePath, err)
			}
		}
	}
}

func safeResumeCacheFilename(resumeID string) string {
	return safeResumeIDPrefix(resumeID) + ".pdf"
}

func safeVersionedResumeCacheFilename(resumeID, versionToken string) string {
	return safeResumeIDPrefix(resumeID) + "_" + versionToken + ".pdf"
}

func safeResumeIDPrefix(resumeID string) string {
	hash := sha256.Sum256([]byte(resumeID))
	return hex.EncodeToString(hash[:])
}

func (h *ResumeHandler) getCacheDir() string {
	if h.CacheDir != "" {
		return h.CacheDir
	}
	return getResumeCacheDir()
}

func (h *ResumeHandler) TailorResume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := h.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	encodedResumeID := r.PathValue("resumeId")
	if encodedResumeID == "" {
		http.Error(w, "resumeId is required", http.StatusBadRequest)
		return
	}

	decodedResumeID, err := h.DecodeID(encodedResumeID)
	if err != nil {
		http.Error(w, "Failed to decode resumeID", http.StatusBadRequest)
		return
	}

	var req struct {
		JobDescription string `json:"jobDescription"`
		Company        string `json:"company"`
		Role           string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.JobDescription == "" {
		http.Error(w, "jobDescription is required", http.StatusBadRequest)
		return
	}

	resume, err := h.Repo.GetResumeByID(r.Context(), decodedResumeID, userID)
	if err != nil {
		http.Error(w, "Resume not found", http.StatusNotFound)
		return
	}

	expInputs := make([]claude.ExperienceInput, len(resume.Experience))
	for i, e := range resume.Experience {
		expInputs[i] = claude.ExperienceInput{
			Employer:     e.Employer,
			Title:        e.Title,
			StartDate:    e.StartDate,
			EndDate:      e.EndDate,
			BulletPoints: e.BulletPoints,
		}
	}

	tailored, err := h.ClaudeClient.TailorResume(r.Context(), claude.TailorRequest{
		Summary:        resume.Summary,
		Skills:         resume.Skills,
		Experience:     expInputs,
		Company:        req.Company,
		Role:           req.Role,
		JobDescription: req.JobDescription,
	})
	if err != nil {
		log.Printf("claude TailorResume error: %v", err)
		http.Error(w, "Failed to tailor resume", http.StatusBadGateway)
		return
	}

	if len(tailored.Experience) != len(resume.Experience) {
		log.Printf("claude returned %d experience entries, expected %d", len(tailored.Experience), len(resume.Experience))
		http.Error(w, "Failed to tailor resume", http.StatusBadGateway)
		return
	}

	resume.Summary = tailored.Summary
	resume.Skills = tailored.Skills
	for i := range resume.Experience {
		resume.Experience[i].BulletPoints = tailored.Experience[i].BulletPoints
	}

	resume.ResumeID = h.EncodeID(resume.ResumeID)
	resume.Profile.ProfileID = h.EncodeID(resume.Profile.ProfileID)

	json.NewEncoder(w).Encode(resume)
}
