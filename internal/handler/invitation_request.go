package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/dfloo/dfloo-profile-go/internal/email"
	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
)

type InvitationRequestHandler struct {
	Repo         repository.InvitationRequestRepository
	EmailService *email.Service
}

func NewInvitationRequestHandler(repo repository.InvitationRequestRepository, emailSvc *email.Service) *InvitationRequestHandler {
	return &InvitationRequestHandler{Repo: repo, EmailService: emailSvc}
}

func (h *InvitationRequestHandler) PostInvitationRequest(w http.ResponseWriter, r *http.Request) {
	var req model.InvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Reason == "" {
		http.Error(w, "Name, email, and reason are required", http.StatusBadRequest)
		return
	}

	if err := h.Repo.CreateInvitationRequest(r.Context(), &req); err != nil {
		http.Error(w, "Failed to save invitation request", http.StatusInternalServerError)
		return
	}

	if err := h.EmailService.Send(
		fmt.Sprintf("Invitation Request from %s", html.EscapeString(req.Name)),
		fmt.Sprintf(
			"<p><strong>Name:</strong> %s</p><p><strong>Email:</strong> %s</p><p><strong>Reason:</strong></p><p>%s</p>",
			html.EscapeString(req.Name), html.EscapeString(req.Email), html.EscapeString(req.Reason),
		),
	); err != nil {
		log.Printf("failed to send invitation request email: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
}
