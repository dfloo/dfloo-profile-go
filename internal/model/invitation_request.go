package model

import "time"

type InvitationRequest struct {
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Reason  string    `json:"reason"`
	Created time.Time `json:"created"`
}
