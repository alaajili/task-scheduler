package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/alaajili/task-scheduler/shared/database"
	"github.com/alaajili/task-scheduler/shared/models"
)

type TaskRepository struct {
	db *database.DB
}

func NewTaskRepository(db *database.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) DB() *database.DB {
	return r.db
}

// CreateTask inserts a new task into the database.
func (r *TaskRepository) CreateTask(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (
			id, type, payload, priority, state,
			retry_count, max_retries, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.Type, task.Payload, task.Priority, task.State,
		task.RetryCount, task.MaxRetries, task.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

// GetTaskByID retrieves a single task by its ID.
func (r *TaskRepository) GetTaskByID(ctx context.Context, id string) (*models.Task, error) {
	query := `
		SELECT id, type, payload, priority, state, result, error,
		       retry_count, max_retries, created_at, started_at,
		       completed_at, worker_id
		FROM tasks
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	task, err := r.scanTask(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// ListTasks retrieves tasks with optional filters, ordering, and pagination.
func (r *TaskRepository) ListTasks(ctx context.Context, filters ListFilters) ([]*models.Task, error) {
	var (
		queryBuilder strings.Builder
		args         []any
		argIndex     = 1
	)

	queryBuilder.WriteString(`
		SELECT id, type, payload, priority, state, result, error,
		       retry_count, max_retries, created_at, started_at,
		       completed_at, worker_id
		FROM tasks
		WHERE 1=1
	`)

	if filters.State != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND state = $%d", argIndex))
		args = append(args, filters.State)
		argIndex++
	}
	if filters.Type != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND type = $%d", argIndex))
		args = append(args, filters.Type)
		argIndex++
	}

	queryBuilder.WriteString(" ORDER BY priority DESC, created_at ASC")

	if filters.Limit > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d", argIndex))
		args = append(args, filters.Limit)
		argIndex++
	}
	if filters.Offset > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" OFFSET $%d", argIndex))
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTaskState updates the state of a task by ID.
func (r *TaskRepository) UpdateTaskState(ctx context.Context, id string, state models.TaskState) error {
	query := `UPDATE tasks SET state = $1 WHERE id = $2`

	res, err := r.db.ExecContext(ctx, query, state, id)
	if err != nil {
		return fmt.Errorf("failed to update task state: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found: %s", id)
	}
	return nil
}

// scanTask maps SQL row data into a Task struct, handling nullable fields.
func (r *TaskRepository) scanTask(scanner interface {
	Scan(dest ...any) error
}) (*models.Task, error) {
	var (
		task        models.Task
		startedAt   sql.NullTime
		completedAt sql.NullTime
		result      sql.NullString
		errorMsg    sql.NullString
		workerID    sql.NullString
	)

	err := scanner.Scan(
		&task.ID, &task.Type, &task.Payload, &task.Priority, &task.State,
		&result, &errorMsg, &task.RetryCount, &task.MaxRetries,
		&task.CreatedAt, &startedAt, &completedAt, &workerID,
	)
	if err != nil {
		return nil, err
	}

	if result.Valid {
		task.Result = []byte(result.String)
	}
	if errorMsg.Valid {
		task.Error = errorMsg.String
	}
	if workerID.Valid {
		task.WorkerID = workerID.String
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return &task, nil
}

// ListFilters defines optional filters for listing tasks.
type ListFilters struct {
	State  models.TaskState
	Type   models.TaskType
	Limit  int
	Offset int
}
