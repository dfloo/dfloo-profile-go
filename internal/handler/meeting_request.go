package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
)

type MeetingRequestHandler struct {
	Repo      repository.MeetingRequestRepository
	SendEmail func(subject, html string) error
}

func NewMeetingRequestHandler(repo repository.MeetingRequestRepository, sendEmail func(string, string) error) *MeetingRequestHandler {
	return &MeetingRequestHandler{Repo: repo, SendEmail: sendEmail}
}

func (h *MeetingRequestHandler) PostMeetingRequest(w http.ResponseWriter, r *http.Request) {
	var req model.MeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Message == "" {
		http.Error(w, "Name, email, and message are required", http.StatusBadRequest)
		return
	}

	if err := h.Repo.CreateMeetingRequest(r.Context(), &req); err != nil {
		http.Error(w, "Failed to save meeting request", http.StatusInternalServerError)
		return
	}

	if err := h.SendEmail(
		fmt.Sprintf("Meeting Request from %s", html.EscapeString(req.Name)),
		fmt.Sprintf(
			"<p><strong>Name:</strong> %s</p><p><strong>Email:</strong> %s</p><p><strong>Message:</strong></p><p>%s</p>",
			html.EscapeString(req.Name), html.EscapeString(req.Email), html.EscapeString(req.Message),
		),
	); err != nil {
		log.Printf("failed to send meeting request email: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
}
