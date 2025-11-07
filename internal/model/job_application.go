package model

import "time"

type JobApplication struct {
	JobApplicationID string    `json:"jobApplicationId"`
	ResumeID         string    `json:"resumeId"`
	Status           string    `json:"status"`
	SortIndex        int       `json:"sortIndex"`
	Company          string    `json:"company"`
	Role             string    `json:"role"`
	Description      string    `json:"description"`
	Notes            string    `json:"notes"`
	Created          time.Time `json:"created"`
	Updated          time.Time `json:"updated"`
}
