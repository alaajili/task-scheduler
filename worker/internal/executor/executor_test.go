package executor_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/alaajili/task-scheduler/shared/models"
	"github.com/alaajili/task-scheduler/worker/internal/executor"
)

func TestExecutor_HTTPRequest(t *testing.T) {
	exec := executor.NewExecutor("test-worker")
	ctx := context.Background()

	payload := json.RawMessage(`{
		"method": "GET",
		"url": "https://httpbin.org/get",
		"timeout": 10
	}`)

	task := &models.Task{
		ID:      "test-1",
		Type:    models.TaskTypeHTTPRequest,
		Payload: payload,
	}

	err := exec.ExecuteTask(ctx, task)
	require.NoError(t, err)
	assert.NotNil(t, task.Result)

	var result map[string]any
	err = json.Unmarshal(task.Result, &result)
	require.NoError(t, err)
	assert.Equal(t, float64(200), result["status_code"])
}

func TestExecutor_LongRunning(t *testing.T) {
	exec := executor.NewExecutor("test-worker")
	ctx := context.Background()

	payload := json.RawMessage(`{
		"duration_seconds": 2,
		"step_count": 4
	}`)

	task := &models.Task{
		ID:      "test-2",
		Type:    models.TaskTypeLongRunning,
		Payload: payload,
	}

	start := time.Now()
	err := exec.ExecuteTask(ctx, task)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, task.Result)
	assert.GreaterOrEqual(t, duration, 2*time.Second)

	var result map[string]any
	err = json.Unmarshal(task.Result, &result)
	require.NoError(t, err)
	assert.True(t, result["completed"].(bool))
	assert.Equal(t, float64(4), result["steps_executed"])
}

func TestExecutor_EmailSend(t *testing.T) {
	exec := executor.NewExecutor("test-worker")
	ctx := context.Background()

	payload := json.RawMessage(`{
		"to": "test@example.com",
		"subject": "Test",
		"body": "Test email"
	}`)

	task := &models.Task{
		ID:      "test-3",
		Type:    models.TaskTypeEmailSend,
		Payload: payload,
	}

	err := exec.ExecuteTask(ctx, task)
	require.NoError(t, err)
	assert.NotNil(t, task.Result)

	var result map[string]any
	err = json.Unmarshal(task.Result, &result)
	require.NoError(t, err)
	assert.True(t, result["sent"].(bool))
	assert.NotEmpty(t, result["message_id"])
}

func TestExecutor_DataProcessing(t *testing.T) {
	exec := executor.NewExecutor("test-worker")
	ctx := context.Background()

	payload := json.RawMessage(`{
		"operation": "aggregate",
		"data": [{"id": 1}, {"id": 2}, {"id": 3}],
		"options": {}
	}`)

	task := &models.Task{
		ID:      "test-4",
		Type:    models.TaskTypeDataProcessing,
		Payload: payload,
	}

	err := exec.ExecuteTask(ctx, task)
	require.NoError(t, err)
	assert.NotNil(t, task.Result)

	var result map[string]any
	err = json.Unmarshal(task.Result, &result)
	require.NoError(t, err)
	assert.Equal(t, float64(3), result["records_count"])
}

func TestExecutor_InvalidTaskType(t *testing.T) {
	exec := executor.NewExecutor("test-worker")
	ctx := context.Background()

	task := &models.Task{
		ID:      "test-5",
		Type:    models.TaskType("invalid_type"),
		Payload: json.RawMessage(`{}`),
	}

	err := exec.ExecuteTask(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no handler registered")
}

func TestExecutor_Timeout(t *testing.T) {
	exec := executor.NewExecutor("test-worker")
	
	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	payload := json.RawMessage(`{
		"duration_seconds": 10,
		"step_count": 5
	}`)

	task := &models.Task{
		ID:      "test-6",
		Type:    models.TaskTypeLongRunning,
		Payload: payload,
	}

	err := exec.ExecuteTask(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}
