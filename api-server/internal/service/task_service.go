package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alaajili/task-scheduler/api-server/internal/repository"
	"github.com/alaajili/task-scheduler/shared/logger"
	"github.com/alaajili/task-scheduler/shared/models"
	"go.uber.org/zap"
)

type TaskService struct {
	repo *repository.TaskRepository
}

func NewTaskService(repo *repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(ctx context.Context, req CreateTaskRequest) (*models.Task, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	task := models.NewTask(req.Type, req.Payload, req.Priority)
	task.MaxRetries = req.MaxRetries

	// Save task to database
	if err := s.repo.CreateTask(ctx, task); err != nil {
		logger.Error("Failed to create task",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	logger.Info("Task created successfully",
		zap.String("task_id", task.ID),
		zap.String("type", string(task.Type)),
		zap.Int("priority", task.Priority),
	)

	// TODO: publish to the queue (redis queue)

	return task, nil
}

func (s *TaskService) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		logger.Error("Failed to get task",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		return nil, err
	}
	return task, nil
}

func (s *TaskService) ListTasks(ctx context.Context, filter repository.ListFilters) ([]*models.Task, error) {
	tasks, err := s.repo.ListTasks(ctx, filter)
	if err != nil {
		logger.Error("Failed to list tasks",
			zap.Error(err),
		)
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) CancelTask(ctx context.Context, taskID string) error {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		logger.Error("Failed to get task for cancellation",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		return err
	}

	if task.State != models.TaskStatePending && task.State != models.TaskStateRunning {
		logger.Warn("Attempted to cancel a task that is not pending or running",
			zap.String("task_id", taskID),
			zap.String("state", string(task.State)),
		)
		return fmt.Errorf("only pending or running tasks can be cancelled")
	}

	if err := s.repo.UpdateTaskState(ctx, taskID, models.TaskStateCancelled); err != nil {
		logger.Error("Failed to cancel task",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		return err
	}

	logger.Info("Task cancelled successfully",
		zap.String("task_id", taskID),
	)
	return nil
}

type CreateTaskRequest struct {
	Type       models.TaskType `json:"type" binding:"required"`
	Payload    json.RawMessage `json:"payload" binding:"required"`
	Priority   int             `json:"priority"`
	MaxRetries int             `json:"max_retries"`
}

func (r *CreateTaskRequest) Validate() error {
	validTypes := map[models.TaskType]bool{
		models.TaskTypeHTTPRequest:    true,
		models.TaskTypeDataProcessing: true,
		models.TaskTypeEmailSend:      true,
		models.TaskTypeLongRunning:    true,
	}

	if !validTypes[r.Type] {
		return fmt.Errorf("invalid task type: %s", r.Type)
	}
	if r.Priority < 0 || r.Priority > 10 {
		return fmt.Errorf("priority must be between 0 and 10")
	}
	if r.MaxRetries < 0 {
		r.MaxRetries = 3 // Default to 3 retries
	}
	return nil
}
