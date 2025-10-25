package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWorker(t *testing.T) {
	taskTypes := []TaskType{TaskTypeHTTPRequest, TaskTypeDataProcessing}
	worker := NewWorker(taskTypes)

	assert.Equal(t, WorkerStatusIdle, worker.Status)
	assert.Equal(t, taskTypes, worker.TaskTypes)
	assert.NotEmpty(t, worker.ID)
	assert.NotZero(t, worker.CreatedAt)
	assert.NotZero(t, worker.LastHeartbeat)
}

func TestCanHandleTask(t *testing.T) {
	taskTypes := []TaskType{TaskTypeHTTPRequest, TaskTypeDataProcessing}
	worker := NewWorker(taskTypes)

	assert.True(t, worker.CanHandleTask(TaskTypeHTTPRequest))
	assert.True(t, worker.CanHandleTask(TaskTypeDataProcessing))
	assert.False(t, worker.CanHandleTask(TaskTypeEmailSend))
	assert.False(t, worker.CanHandleTask(TaskTypeLongRunning))
}

func TestIsHealthy(t *testing.T) {
	worker := NewWorker([]TaskType{TaskTypeHTTPRequest})

	assert.True(t, worker.IsHealthy(5*time.Second))
	
	// Simulate a heartbeat delay
	worker.LastHeartbeat = worker.LastHeartbeat.Add(-10 * time.Second)
	assert.False(t, worker.IsHealthy(5*time.Second))
}

func TestMarkHeartbeat(t *testing.T) {
	worker := NewWorker([]TaskType{TaskTypeHTTPRequest})
	originalHeartbeat := worker.LastHeartbeat

	time.Sleep(3 * time.Second)
	worker.MarkHeartbeat()

	assert.True(t, worker.LastHeartbeat.After(originalHeartbeat))
}

