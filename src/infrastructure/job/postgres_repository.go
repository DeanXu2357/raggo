package job

import (
	"context"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type PostgresJobRepository struct {
	db *gorm.DB
}

func NewPostgresJobRepository(db *gorm.DB) *PostgresJobRepository {
	return &PostgresJobRepository{db: db}
}

func (r *PostgresJobRepository) Create(ctx context.Context, taskType string, payload json.RawMessage) (*Job, error) {
	job := &Job{
		TaskType: taskType,
		Payload:  payload,
		Status:   JobStatusPending,
	}

	result := r.db.WithContext(ctx).Create(job)
	if result.Error != nil {
		return nil, result.Error
	}

	return job, nil
}

func (r *PostgresJobRepository) Get(ctx context.Context, id int) (*Job, error) {
	var job Job
	result := r.db.WithContext(ctx).First(&job, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &job, nil
}

func (r *PostgresJobRepository) UpdateStatus(ctx context.Context, id int, status JobStatus, err *string) error {
	result := r.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": status,
		"error":  err,
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("job not found")
	}

	return nil
}
