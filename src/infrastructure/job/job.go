package job

import (
	"context"
	"encoding/json"
	"time"
)

// JobStatus defines the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// Job represents a background job
type Job struct {
	ID        int             `json:"id"`
	TaskType  string          `json:"task_type"`
	Payload   json.RawMessage `json:"payload"`
	Status    JobStatus       `json:"status"`
	Error     *string         `json:"error,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// JobRepository defines the interface for job persistence
type JobRepository interface {
	Create(ctx context.Context, taskType string, payload json.RawMessage) (*Job, error)
	Get(ctx context.Context, id int) (*Job, error)
	UpdateStatus(ctx context.Context, id int, status JobStatus, err *string) error
}
