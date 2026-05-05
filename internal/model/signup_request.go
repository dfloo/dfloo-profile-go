package model

import "time"

type SignupRequest struct {
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Reason  string    `json:"reason"`
	Created time.Time `json:"created"`
}
