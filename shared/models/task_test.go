package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTask(t *testing.T) {
	payload := json.RawMessage(`{"url": "https://example.com"}`)
	task := NewTask(TaskTypeHTTPRequest, payload, 5)

	assert.Equal(t, TaskTypeHTTPRequest, task.Type)
	assert.Equal(t, payload, task.Payload)
	assert.Equal(t, 5, task.Priority)
	assert.Equal(t, TaskStatePending, task.State)
	assert.Equal(t, 0, task.RetryCount)
	assert.Equal(t, 3, task.MaxRetries)
	assert.NotEmpty(t, task.ID)
	assert.NotZero(t, task.CreatedAt)
}

func TestTaskPriorityBounds(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{-1, 0},
		{0, 0},
		{5, 5},
		{10, 10},
		{15, 10},
	}

	for _, tt := range tests {
		task := NewTask(TaskTypeDataProcessing, nil, tt.input)
		assert.Equal(t, tt.expected, task.Priority)
	}
}

func TestTaskCanRetry(t *testing.T) {
	task := NewTask(TaskTypeEmailSend, nil, 3)
	task.MaxRetries = 2

	assert.True(t, task.CanRetry())

	task.RetryCount = 2
	assert.False(t, task.CanRetry())
}

func TestTaskLifecycle(t *testing.T) {
	task := NewTask(TaskTypeHTTPRequest, nil, 5)
	workerID := "worker-123"

	task.MarkStarted(workerID)
	assert.Equal(t, TaskStateRunning, task.State)
	assert.NotNil(t, task.StartedAt)
	assert.Equal(t, workerID, task.WorkerID)

	result := json.RawMessage(`{"status": "success"}`)
	task.MarkCompleted(result)
	assert.Equal(t, TaskStateCompleted, task.State)
	assert.NotNil(t, task.CompletedAt)
	assert.Equal(t, result, task.Result)
}
