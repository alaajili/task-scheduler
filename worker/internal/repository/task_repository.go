package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/alaajili/task-scheduler/shared/database"
	"github.com/alaajili/task-scheduler/shared/models"
)

// TaskRepository handles database operations for tasks
type TaskRepository struct {
	db *database.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *database.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// GetNextPendingTask retrieves the next pending task with priority
func (r *TaskRepository) GetNextPendingTask(ctx context.Context, taskTypes []models.TaskType) (*models.Task, error) {
	query := `
		SELECT id, type, payload, priority, state, 
		       retry_count, max_retries, created_at
		FROM tasks
		WHERE state = 'pending'
		  AND type = ANY($1)
		ORDER BY priority DESC, created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`

	// Convert TaskType slice to string slice
	typeStrings := make([]string, len(taskTypes))
	for i, t := range taskTypes {
		typeStrings[i] = string(t)
	}

	var task models.Task
	// Use pq.Array to convert to PostgreSQL array type
	err := r.db.QueryRowContext(ctx, query, pq.Array(typeStrings)).Scan(
		&task.ID, &task.Type, &task.Payload, &task.Priority, &task.State,
		&task.RetryCount, &task.MaxRetries, &task.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No tasks available
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get next task: %w", err)
	}

	return &task, nil
}

// MarkTaskStarted marks a task as started
func (r *TaskRepository) MarkTaskStarted(ctx context.Context, taskID, workerID string) error {
	query := `
		UPDATE tasks 
		SET state = 'running', 
		    started_at = NOW(), 
		    worker_id = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, taskID, workerID)
	if err != nil {
		return fmt.Errorf("failed to mark task as started: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found or already started: %s", taskID)
	}

	return nil
}

// MarkTaskCompleted marks a task as completed with result
func (r *TaskRepository) MarkTaskCompleted(ctx context.Context, taskID string, result json.RawMessage) error {
	query := `
		UPDATE tasks 
		SET state = 'completed', 
		    completed_at = NOW(), 
		    result = $2
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, taskID, result)
	if err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	return nil
}

// MarkTaskFailed marks a task as failed with error
func (r *TaskRepository) MarkTaskFailed(ctx context.Context, taskID, errorMsg string) error {
	query := `
		UPDATE tasks 
		SET state = 'failed', 
		    completed_at = NOW(), 
		    error = $2,
		    retry_count = retry_count + 1
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, taskID, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to mark task as failed: %w", err)
	}

	return nil
}

// MarkTaskForRetry resets a failed task to pending for retry
func (r *TaskRepository) MarkTaskForRetry(ctx context.Context, taskID string, retryDelay time.Duration) error {
	query := `
		UPDATE tasks 
		SET state = 'pending',
		    started_at = NULL,
		    worker_id = NULL
		WHERE id = $1
		  AND retry_count < max_retries
	`

	result, err := r.db.ExecContext(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("failed to mark task for retry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not eligible for retry: %s", taskID)
	}

	return nil
}

// GetTaskByID retrieves a task by ID
func (r *TaskRepository) GetTaskByID(ctx context.Context, taskID string) (*models.Task, error) {
	query := `
		SELECT id, type, payload, priority, state, result, error,
		       retry_count, max_retries, created_at, started_at, 
		       completed_at, worker_id
		FROM tasks 
		WHERE id = $1
	`

	var task models.Task
	var result, errorMsg sql.NullString
	var startedAt, completedAt sql.NullTime
	var workerID sql.NullString

	err := r.db.QueryRowContext(ctx, query, taskID).Scan(
		&task.ID, &task.Type, &task.Payload, &task.Priority, &task.State,
		&result, &errorMsg, &task.RetryCount, &task.MaxRetries,
		&task.CreatedAt, &startedAt, &completedAt, &workerID,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Handle nullable fields
	if result.Valid {
		task.Result = []byte(result.String)
	}
	if errorMsg.Valid {
		task.Error = errorMsg.String
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if workerID.Valid {
		task.WorkerID = workerID.String
	}

	return &task, nil
}
