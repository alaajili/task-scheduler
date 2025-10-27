package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/alaajili/task-scheduler/shared/config"
	"github.com/alaajili/task-scheduler/shared/database"
	"github.com/alaajili/task-scheduler/shared/models"
)

func TestDB(t *testing.T) *database.DB {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:	  5432,
		User:	  "postgres",
		Password: "postgres",
		DBName:   "taskscheduler_test",
		SSLMode:  "disable",
		MaxConns: 5,
		MaxIdle:  2,	
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("failed to connect to test databse: %v", err)
	}

	t.Cleanup(func() {
		CleanupDB(t, db)
		db.Close()
	})

	return db
}

func CleanupDB(t *testing.T, db *database.DB) {
	ctx := context.Background()

	tables := []string{"workers", "tasks"}
	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}
}

func WaitForCondition(t *testing.T, timeout time.Duration, interval time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

func CreateTestTask(t *testing.T, db *database.DB, task *models.Task) {
	ctx := context.Background()

	query := `
		INSERT INTO tasks (
			id, type, payload, prority, state,
			retry_count, max_retries, creqated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := db.ExecContext(ctx, query,
		task.ID,
		task.Type,
		task.Payload,
		task.Priority,
		task.State,
		task.RetryCount,
		task.MaxRetries,
		task.CreatedAt,
	)

	if err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}
}

func GetTestTaskByID(t *testing.T, db *database.DB, id string) *models.Task {
	ctx := context.Background()

	query := `
		SELECT id, type, payload, priority, state, result, error,
		       retry_count, max_retries, created_at, started_at,
			   completed_at, worker_id
		FROM tasks WHERE id = $1
	`

	var task models.Task
	var result, errorMsg sql.NullString
	var startedAt, completedAt sql.NullTime
	var workerID sql.NullString
	
	err := db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Type, &task.Payload, &task.Priority, &task.State,
		&result, &errorMsg, &task.RetryCount, &task.MaxRetries,
		&task.CreatedAt,&startedAt, &completedAt, &workerID,
	)

	if err != nil {
		t.Fatalf("failed to get test task by ID: %v", err)
	}

	if result.Valid {
		task.Result = []byte(result.String)
	}
	if errorMsg.Valid {
		task.Error = errorMsg.String
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if workerID.Valid {
		task.WorkerID = workerID.String
	}

	return &task
}
