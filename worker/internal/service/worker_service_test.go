package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alaajili/task-scheduler/shared/config"
	"github.com/alaajili/task-scheduler/shared/models"
	"github.com/alaajili/task-scheduler/shared/queue"
	"github.com/alaajili/task-scheduler/shared/testutil"
	"github.com/alaajili/task-scheduler/worker/internal/repository"
	"github.com/alaajili/task-scheduler/worker/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupWorkerTest(t *testing.T) (*service.WorkerService, *repository.TaskRepository) {
	db := testutil.TestDB(t)
	repo := repository.NewTaskRepository(db)
	cfg, _ := config.LoadConfig("")
	rq, _ := queue.NewRedisQueue(cfg.Redis)
	
	taskTypes := []models.TaskType{
		models.TaskTypeHTTPRequest,
		models.TaskTypeDataProcessing,
		models.TaskTypeEmailSend,
		models.TaskTypeLongRunning,
	}
	
	workerService := service.NewWorkerService("test-worker", repo, rq,taskTypes)
	
	return workerService, repo
}

func TestProcessNextTask_Success(t *testing.T) {
	workerService, repo := setupWorkerTest(t)
	ctx := context.Background()

	// Create a test task
	payload := json.RawMessage(`{"duration_seconds": 1, "step_count": 2}`)
	task := models.NewTask(models.TaskTypeLongRunning, payload, 5)
	testutil.CreateTestTask(t, repo.DB(), task)

	// Process the task
	processed, err := workerService.ProcessNextTask(ctx)
	require.NoError(t, err)
	assert.True(t, processed)

	// Verify task is completed
	updatedTask, err := repo.GetTaskByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.TaskStateCompleted, updatedTask.State)
	assert.NotNil(t, updatedTask.Result)
	assert.NotNil(t, updatedTask.StartedAt)
	assert.NotNil(t, updatedTask.CompletedAt)
}

func TestProcessNextTask_NoTasksAvailable(t *testing.T) {
	workerService, _ := setupWorkerTest(t)
	ctx := context.Background()

	// Process when no tasks exist
	processed, err := workerService.ProcessNextTask(ctx)
	require.NoError(t, err)
	assert.False(t, processed)
}

func TestProcessNextTask_TaskFailsAndRetries(t *testing.T) {
	workerService, repo := setupWorkerTest(t)
	ctx := context.Background()

	// Create a task that will fail
	payload := json.RawMessage(`{
		"duration_seconds": 1,
		"step_count": 2,
		"simulate_error": true,
		"error_after": 0
	}`)
	task := models.NewTask(models.TaskTypeLongRunning, payload, 5)
	task.MaxRetries = 3
	testutil.CreateTestTask(t, repo.DB(), task)

	// Process the task (should fail)
	processed, err := workerService.ProcessNextTask(ctx)
	require.NoError(t, err)
	assert.True(t, processed)

	// Verify task is failed
	updatedTask, err := repo.GetTaskByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.TaskStateFailed, updatedTask.State)
	assert.Equal(t, 1, updatedTask.RetryCount)
	assert.NotEmpty(t, updatedTask.Error)
}

func TestProcessNextTask_HTTPRequest(t *testing.T) {
	workerService, repo := setupWorkerTest(t)
	ctx := context.Background()

	// Create HTTP request task
	payload := json.RawMessage(`{
		"method": "GET",
		"url": "https://httpbin.org/get",
		"timeout": 30
	}`)
	task := models.NewTask(models.TaskTypeHTTPRequest, payload, 5)
	testutil.CreateTestTask(t, repo.DB(), task)

	// Process the task
	processed, err := workerService.ProcessNextTask(ctx)
	require.NoError(t, err)
	assert.True(t, processed)

	// Verify task is completed
	updatedTask, err := repo.GetTaskByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.TaskStateCompleted, updatedTask.State)
	
	// Parse result
	var result map[string]any
	err = json.Unmarshal(updatedTask.Result, &result)
	require.NoError(t, err)
	assert.Equal(t, float64(200), result["status_code"])
}

func TestProcessNextTask_EmailSend(t *testing.T) {
	workerService, repo := setupWorkerTest(t)
	ctx := context.Background()

	// Create email task
	payload := json.RawMessage(`{
		"to": "test@example.com",
		"subject": "Test Email",
		"body": "This is a test email"
	}`)
	task := models.NewTask(models.TaskTypeEmailSend, payload, 5)
	testutil.CreateTestTask(t, repo.DB(), task)

	// Process the task
	processed, err := workerService.ProcessNextTask(ctx)
	require.NoError(t, err)
	assert.True(t, processed)

	// Verify task is completed
	updatedTask, err := repo.GetTaskByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.TaskStateCompleted, updatedTask.State)
	
	// Parse result
	var result map[string]any
	err = json.Unmarshal(updatedTask.Result, &result)
	require.NoError(t, err)
	assert.True(t, result["sent"].(bool))
}

func TestProcessNextTask_DataProcessing(t *testing.T) {
	workerService, repo := setupWorkerTest(t)
	ctx := context.Background()

	// Create data processing task
	payload := json.RawMessage(`{
		"operation": "aggregate",
		"data": [
			{"id": 1, "value": 100},
			{"id": 2, "value": 200}
		],
		"options": {}
	}`)
	task := models.NewTask(models.TaskTypeDataProcessing, payload, 5)
	testutil.CreateTestTask(t, repo.DB(), task)

	// Process the task
	processed, err := workerService.ProcessNextTask(ctx)
	require.NoError(t, err)
	assert.True(t, processed)

	// Verify task is completed
	updatedTask, err := repo.GetTaskByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.TaskStateCompleted, updatedTask.State)
	
	// Parse result
	var result map[string]any
	err = json.Unmarshal(updatedTask.Result, &result)
	require.NoError(t, err)
	assert.Equal(t, "aggregate", result["operation"])
	assert.Equal(t, float64(2), result["records_count"])
}

// func TestRetryDelay_ExponentialBackoff(t *testing.T) {
// 	workerService, _ := setupWorkerTest(t)

// 	tests := []struct {
// 		retryCount    int
// 		minDelay      time.Duration
// 		maxDelay      time.Duration
// 	}{
// 		{1, 4*time.Second, 6*time.Second},       // ~5s with jitter
// 		{2, 22*time.Second, 28*time.Second},     // ~25s with jitter
// 		{3, 112*time.Second, 138*time.Second},   // ~125s with jitter
// 		{4, 560*time.Second, 600*time.Second},   // capped at 600s
// 	}

// 	for _, tt := range tests {
// 		t.Run(fmt.Sprintf("retry_%d", tt.retryCount), func(t *testing.T) {
// 			// We can't test private method directly, but we can test via task failure
// 			// This is more of an integration test
			
// 			// For unit testing, we'd need to expose calculateRetryDelay or test via behavior
// 			// For now, we'll verify the behavior through integration tests
// 		})
// 	}
// }
