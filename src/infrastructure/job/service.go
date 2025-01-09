package job

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type JobService struct {
	publisher       message.Publisher
	repo            JobRepository
	logger          watermill.LoggerAdapter
	translationTask *TranslationTask
}

type JobMessage struct {
	JobID    int             `json:"job_id"`
	TaskType string          `json:"task_type"`
	Payload  json.RawMessage `json:"payload"`
}

type TestPayload struct {
	Print string `json:"print"`
}

func NewJobService(
	publisher message.Publisher,
	repo JobRepository,
	logger watermill.LoggerAdapter,
	translator *TranslationTask,
) *JobService {
	return &JobService{
		publisher:       publisher,
		repo:            repo,
		logger:          logger,
		translationTask: translator,
	}
}

// EnqueueJob creates a new job and publishes it to the message queue
func (s *JobService) EnqueueJob(ctx context.Context, taskType string, payload json.RawMessage) (*Job, error) {
	// Create job record
	job, err := s.repo.Create(ctx, taskType, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Prepare message
	jobMsg := JobMessage{
		JobID:    job.ID,
		TaskType: job.TaskType,
		Payload:  job.Payload,
	}

	msgPayload, err := json.Marshal(jobMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job message: %w", err)
	}

	// Publish message
	msg := message.NewMessage(watermill.NewUUID(), msgPayload)
	if err := s.publisher.Publish("jobs", msg); err != nil {
		return nil, fmt.Errorf("failed to publish job message: %w", err)
	}

	return job, nil
}

// ProcessJobMessage processes a job message from the queue
func (s *JobService) ProcessJobMessage(msg *message.Message) error {
	var jobMsg JobMessage
	if err := json.Unmarshal(msg.Payload, &jobMsg); err != nil {
		return fmt.Errorf("failed to unmarshal job message: %w", err)
	}

	ctx := context.Background()

	// Get job from database
	job, err := s.repo.Get(ctx, jobMsg.JobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found: %d", jobMsg.JobID)
	}

	// Update status to running
	if err := s.repo.UpdateStatus(ctx, job.ID, JobStatusRunning, nil); err != nil {
		return fmt.Errorf("failed to update job status to running: %w", err)
	}

	// Process the job based on task type
	err = s.processJob(ctx, job)

	if err != nil {
		// Update status to failed
		errStr := err.Error()
		if updateErr := s.repo.UpdateStatus(ctx, job.ID, JobStatusFailed, &errStr); updateErr != nil {
			s.logger.Error("Failed to update job status to failed", updateErr, watermill.LogFields{
				"job_id": job.ID,
			})
		}
		return fmt.Errorf("failed to process job: %w", err)
	}

	// Update status to completed
	if err := s.repo.UpdateStatus(ctx, job.ID, JobStatusCompleted, nil); err != nil {
		return fmt.Errorf("failed to update job status to completed: %w", err)
	}

	return nil
}

// processJob handles different types of jobs
func (s *JobService) processJob(ctx context.Context, job *Job) error {
	switch job.TaskType {
	case "test":
		var payload TestPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal test payload: %w", err)
		}
		s.logger.Info("Test job executed", watermill.LogFields{
			"job_id": job.ID,
			"print":  payload.Print,
		})
		return nil
	case TaskTypeTranslation:
		return s.translationTask.HandleTranslationTask(ctx, job.Payload)
	default:
		return fmt.Errorf("unknown task type: %s", job.TaskType)
	}
}
