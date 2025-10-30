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
	
	workerID := fmt.Sprintf("worker-%s", uuid.New().String()[:8])
	taskRepo := repository.NewTaskRepository(db)
	
	taskTypes := []models.TaskType{
		models.TaskTypeHTTPRequest,
		models.TaskTypeDataProcessing,
		models.TaskTypeEmailSend,
		models.TaskTypeLongRunning,
	}
	workerService := service.NewWorkerService(workerID, taskRepo, taskTypes)
	logger.Info("New worker started",
		zap.String("worker_id", workerID),
		// zap.Strings("task_types", taskTypesToStrings(taskTypes)),
	)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go workerLoop(ctx, workerService, cfg.Worker.TaskPollInterval)
	
	sig := <-sigChan
	logger.Info("Received Shutdown signal", zap.String("signal", sig.String()))
	
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

func workerLoop(ctx context.Context, s *service.WorkerService, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			logger.Info("Worker Loop stopping")
			return
		case <-ticker.C:
			processed, err := s.ProcessNextTask(ctx)
			if err != nil {
				logger.Error("Error processing task", zap.Error(err))
				continue
			}
			if !processed {
				// no tasks available, wait for the next thick
				continue
			}
			
			// task was  processed, check immmediatly for the next one 
			ticker.Reset(pollInterval)
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
