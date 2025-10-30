package service

import (
	"context"
	"math/rand"
	"fmt"
	"math"
	"time"

	"github.com/alaajili/task-scheduler/shared/logger"
	"github.com/alaajili/task-scheduler/shared/models"
	"github.com/alaajili/task-scheduler/worker/internal/executor"
	"github.com/alaajili/task-scheduler/worker/internal/repository"
	"go.uber.org/zap"
)


type WorkerService struct {
	workerID string
	taskRepo   *repository.TaskRepository
	executor   *executor.Executor
	taskTypes  []models.TaskType
	maxRetries int
}

func NewWorkerService(
	workerID  string,
	taskRepo  *repository.TaskRepository,
	taskTypes []models.TaskType,
) *WorkerService {
	return &WorkerService{
		workerID: workerID,
		taskRepo: taskRepo,
		executor: executor.NewExecutor(workerID),
		taskTypes: taskTypes,
		maxRetries: 5,
	}
}

// ProcessNextTask fetches and process the next available task
func (s *WorkerService) ProcessNextTask(ctx context.Context) (bool, error) {
	task, err := s.taskRepo.GetNextPendingTask(ctx, s.taskTypes)
	if err != nil {
		return false, fmt.Errorf("failed to get next pending task: %w", err)
	}
	
	// No task pending available
	if task == nil {
		return false, nil
	}
	
	logger.Info("Processing task",
		zap.String("task_id", task.ID),
		zap.String("task_type", string(task.Type)),
		zap.Int("priority", task.Priority),
	)
	
	if err := s.taskRepo.MarkTaskStarted(ctx, task.ID, s.workerID); err != nil {
		logger.Error("Failed to mark task as started",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return false, err
	}
	
	// execute task
	err = s.executor.ExecuteTask(ctx, task)
	if err != nil {
		logger.Error("Task execution failed",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		
		if err := s.handleTaskFailure(ctx, task, err); err != nil {
			logger.Error("Failed to handle task failure",
				zap.String("task_id", task.ID),
				zap.Error(err),
			)
		}
		return true, nil
	}
	
	if err := s.taskRepo.MarkTaskCompleted(ctx, task.ID, task.Result); err != nil {
		logger.Error("Failed to mark test as completed",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return true, err
	}
	
	logger.Info("Task completed succefully",
		zap.String("task_id", task.ID),
	)
	
	return true, nil
}

func (s *WorkerService) handleTaskFailure(ctx context.Context, task *models.Task, execErr error) error {
	if err := s.taskRepo.MarkTaskFailed(ctx, task.ID, execErr.Error()); err != nil {
		return err
	}
	
	if task.RetryCount+1 < task.MaxRetries {
		delay := s.calculateRetryDelay(task.RetryCount + 1)
		logger.Info("Scheduling task for retry",
			zap.String("task_id", task.ID),
			zap.Int("retry_count", task.RetryCount+1),
			zap.Duration("retry_delay", delay),
		)
		
		time.Sleep(delay)
		if err := s.taskRepo.MarkTaskForRetry(ctx, task.ID, delay); err != nil {
			return err
		}
	} else {
		logger.Warn("Task exceeded max retries",
			zap.String("task_id", task.ID),
			zap.Int("retry_count", task.RetryCount+1),
		)
	}
	return nil
}

func (s *WorkerService) calculateRetryDelay(retryCount int) time.Duration {
	baseDelay := 5.0
	
	// Exponential backoff: base^retryCount seconds
	delay := math.Pow(baseDelay, float64(retryCount))
	
	// cap at 10 min
	maxDelay := 600.0
	if delay > maxDelay {
		delay = maxDelay
	}
	
	// add jitter (±10%)
	jitter := delay * 0.1 * (rand.Float64()*2 - 1)
	delay += jitter
	
	return time.Duration(delay) * time.Second
}

func (s *WorkerService) GetSupportedTaskTypes() []models.TaskType {
	return s.taskTypes
}
