package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alaajili/task-scheduler/shared/logger"
	"github.com/alaajili/task-scheduler/shared/models"
	"go.uber.org/zap"
)


type Executor struct {
	workerID string
	handlers map[models.TaskType]TaskHandler
}

type TaskHandler func(ctx context.Context, payload json.RawMessage) (json.RawMessage, error)

func NewExecutor(workerID string) *Executor {
	e := &Executor{
		workerID: workerID,
		handlers: make(map[models.TaskType]TaskHandler),
	}

	// register task handlers
	e.RegisterHandler(models.TaskTypeHTTPRequest, e.executeHTTPRequest)
	e.RegisterHandler(models.TaskTypeDataProcessing, e.executeDataProcessing)
	e.RegisterHandler(models.TaskTypeEmailSend, e.executeEmailSend)
	e.RegisterHandler(models.TaskTypeLongRunning, e.executeLongRunning)

	return e
}

func (e *Executor) RegisterHandler(taskType models.TaskType, handler TaskHandler) {
	e.handlers[taskType] = handler
}

func (e *Executor) ExecuteTask(ctx context.Context, task *models.Task) error {
	logger.Info("Executing task",
		zap.String("task_id", task.ID),
		zap.String("task_type", string(task.Type)),
		zap.String("worker_id", e.workerID),
	)

	hanler, exists := e.handlers[task.Type]
	if !exists {
		logger.Error("No handler registered for task type",
			zap.String("task_type", string(task.Type)),
		)
		return fmt.Errorf("no handler registered for task type: %s", task.Type)
	}
	
	execCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	startTime := time.Now()
	result, err := hanler(execCtx, task.Payload)
	duration := time.Since(startTime)

	logger.Info("Task execution completed",
		zap.String("task_id", task.ID),
		zap.String("worker_id", e.workerID),
		zap.Duration("duration", duration),
		zap.Bool("success", err == nil),
	)
	if err != nil {
		return fmt.Errorf("task execution failed: %w", err)
	}

	task.Result = result
	return nil
}

func (e *Executor) CanHandle(taskType models.TaskType) bool {
	_, exists := e.handlers[taskType]
	return exists
}
