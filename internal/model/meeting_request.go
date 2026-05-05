package model

import "time"

type MeetingRequest struct {
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Message string    `json:"message"`
	Created time.Time `json:"created"`
}
