package models

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

// WorkerStatus represents the current status of a worker
type WorkerStatus string

const (
	WorkerStatusActive   WorkerStatus = "ACTIVE"
	WorkerStatusIdle     WorkerStatus = "IDLE"
	WorkerStatusShutdown WorkerStatus = "SHUTDOWN"
)

// Worker represents a worker node in the system.
type Worker struct {
	ID            string       `json:"id" db:"id"`
	Status	      WorkerStatus `json:"status" db:"status"`
	CurrentTaskID string       `json:"current_task_id,omitempty" db:"current_task_id"`
	LastHeartbeat time.Time    `json:"last_heartbeat" db:"last_heartbeat"`
	TaskTypes     []TaskType `json:"task_types" db:"task_types"` // Types of tasks this worker can handle
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
}

func NewWorker(taskTypes []TaskType) *Worker {
	return &Worker{
		ID:            uuid.New().String(),
		Status:        WorkerStatusIdle,
		LastHeartbeat: time.Now().UTC(),
		TaskTypes:     taskTypes,
		CreatedAt:     time.Now().UTC(),
	}
}

func (w *Worker) CanHandleTask(taskType TaskType) bool {
	return slices.Contains(w.TaskTypes, taskType)
}

func (w *Worker) IsHealthy(threshold time.Duration) bool {
	return time.Since(w.LastHeartbeat) < threshold
}

func (w *Worker) MarkHeartbeat() {
	w.LastHeartbeat = time.Now().UTC()
}
