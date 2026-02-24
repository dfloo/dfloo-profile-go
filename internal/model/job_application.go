package model

import (
	"encoding/json"
	"time"
)

type JobApplication struct {
	JobApplicationID string          `json:"jobApplicationId"`
	ResumeID         string          `json:"resumeId"`
	Status           string          `json:"status"`
	SortIndex        int             `json:"sortIndex"`
	Company          string          `json:"company"`
	Role             string          `json:"role"`
	Description      string          `json:"description"`
	Notes            string          `json:"notes"`
	SourceURL        string          `json:"sourceUrl,omitempty"`
	Snapshot         json.RawMessage `json:"snapshot,omitempty"`
	Created          time.Time       `json:"created"`
	Updated          time.Time       `json:"updated"`
}

type FromURLRequest struct {
	URL string `json:"url"`
}
