package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alaajili/task-scheduler/shared/config"
	"github.com/alaajili/task-scheduler/shared/database"
	"github.com/alaajili/task-scheduler/shared/logger"
	"github.com/alaajili/task-scheduler/shared/models"
	"github.com/alaajili/task-scheduler/shared/queue"
	"github.com/alaajili/task-scheduler/worker/internal/repository"
	"github.com/alaajili/task-scheduler/worker/internal/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	if err := logger.Init("development"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	
	cfg, err := config.LoadConfig("../config.yaml")
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}
	
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to the database", zap.Error(err))
	}
	defer db.Close()
	
	rq, err := queue.NewRedisQueue(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to redis", zap.Error(err))
	}
	defer rq.Close()
	
	workerID := fmt.Sprintf("worker-%s", uuid.New().String()[:8])
	taskRepo := repository.NewTaskRepository(db)
	
	taskTypes := []models.TaskType{
		models.TaskTypeHTTPRequest,
		models.TaskTypeDataProcessing,
		models.TaskTypeEmailSend,
		models.TaskTypeLongRunning,
	}
	workerService := service.NewWorkerService(workerID, taskRepo, rq, taskTypes)
	logger.Info("New worker started",
		zap.String("worker_id", workerID),
		zap.Strings("task_types", taskTypesToStrings(taskTypes)),
	)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go workerLoop(ctx, workerService, cfg.Worker.TaskPollInterval)
	
	sig := <-sigChan
	logger.Info("Received Shutdown signal", zap.String("signal", sig.String()))
	
	cancel()
	
	shutdownTimeout := cfg.Worker.GracefulShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = 30 * time.Second
	}
	logger.Info("Waiting for tasks to be completed",
		zap.Duration("timeout", shutdownTimeout),
	)
	
	time.Sleep(2 * time.Second)
	
	logger.Info("Worker stopped", zap.String("worker_id", workerID))
}

func workerLoop(ctx context.Context, workerService *service.WorkerService, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	consecutiveEmptyPolls := 0
	maxBackoff := 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			logger.Info("Worker loop stopping")
			return
		case <-ticker.C:
			// Process next task
			processed, err := workerService.ProcessNextTask(ctx)
			if err != nil {
				logger.Error("Error processing task", zap.Error(err))
				continue
			}

			if !processed {
				// No tasks available - implement backoff
				consecutiveEmptyPolls++
				backoff := time.Duration(consecutiveEmptyPolls) * pollInterval
				backoff = min(backoff, maxBackoff)
				ticker.Reset(backoff)
			} else {
				// Task was processed - reset backoff
				consecutiveEmptyPolls = 0
				ticker.Reset(pollInterval)
			}
		}
	}
}

func taskTypesToStrings(taskTypes []models.TaskType) []string {
	result := make([]string, len(taskTypes))
	for i, t := range taskTypes {
		result[i] = string(t)
	}
	return result
}
