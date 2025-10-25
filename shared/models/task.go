package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TaskState represents the state of a task in the system.
type TaskState string

const (
	TaskStatePending   TaskState = "PENDING"
	TaskStateRunning   TaskState = "RUNNING"
	TaskStateCompleted TaskState = "COMPLETED"
	TaskStateFailed    TaskState = "FAILED"
	TaskStateCanceled  TaskState = "CANCELED"
)

// TaskType represents the type of a task to be executed.
type TaskType string

const (
	TaskTypeHTTPRequest    TaskType = "HTTP_REQUEST"
	TaskTypeDataProcessing TaskType = "DATA_PROCESSING"
	TaskTypeEmailSend      TaskType = "EMAIL_SEND"
	TaskTypeLongRunning    TaskType = "LONG_RUNNING"
)

// Task represents a task in the system.
type Task struct {
	ID          string          `json:"id" db:"id"`
	Type        TaskType        `json:"type" db:"type"`
	Payload     json.RawMessage `json:"payload" db:"payload"`
	Priority    int             `json:"priority" db:"priority"` // 0-10, higher = more urgent
	State       TaskState       `json:"state" db:"state"`
	Result      json.RawMessage `json:"result,omitempty" db:"result"`
	Error       string          `json:"error,omitempty" db:"error"`
	RetryCount  int             `json:"retry_count" db:"retry_count"`
	MaxRetries  int             `json:"max_retries" db:"max_retries"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	WorkerID    string          `json:"worker_id,omitempty" db:"worker_id"`
}

func NewTask(taskType TaskType, payload json.RawMessage, priority int) *Task {
	if priority < 0 {
		priority = 0
	} else if priority > 10 {
		priority = 10
	}

	return &Task{
		ID:         uuid.New().String(),
		Type:       taskType,
		Payload:    payload,
		Priority:   priority,
		State:      TaskStatePending,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now().UTC(),
	}
}

func (t *Task) CanRetry() bool {
	return t.RetryCount < t.MaxRetries
}

func (t *Task) MarkStarted(workerID string) {
	now := time.Now().UTC()
	t.State = TaskStateRunning
	t.StartedAt = &now
	t.WorkerID = workerID
}

func (t *Task) MarkCompleted(result json.RawMessage) {
	now := time.Now().UTC()
	t.State = TaskStateCompleted
	t.Result = result
	t.CompletedAt = &now
}

func (t *Task) MarkFailed(err error) {
	now := time.Now().UTC()
	t.State = TaskStateFailed
	t.Error = err.Error()
	t.CompletedAt = &now
	t.RetryCount++
}

func (t *Task) Duration() time.Duration {
	if t.StartedAt == nil || t.CompletedAt == nil {
		return 0
	}
	return t.CompletedAt.Sub(*t.StartedAt)
}
