package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alaajili/task-scheduler/api-server/internal/handlers"
	"github.com/alaajili/task-scheduler/api-server/internal/repository"
	"github.com/alaajili/task-scheduler/api-server/internal/service"
	"github.com/alaajili/task-scheduler/shared/models"
	"github.com/alaajili/task-scheduler/shared/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *repository.TaskRepository) {
	gin.SetMode(gin.TestMode)

	db := testutil.TestDB(t)
	repository := repository.NewTaskRepository(db)
	service := service.NewTaskService(repository)
	handler := handlers.NewTaskHandler(service)

	router := gin.New()
	apiV1 := router.Group("/api/v1")
	{
		tasks := apiV1.Group("/tasks")
		{
			tasks.POST("", handler.CreateTask)
			tasks.GET("/:id", handler.GetTask)
			tasks.GET("", handler.ListTasks)
			tasks.DELETE("/:id", handler.CancelTask)
		}
	}
	return router, repository
}

func TestCreateTask(t *testing.T) {
	router, _ := setupTestRouter(t)

	payload := map[string]any{
		"url":      "http://example.com",
		"method":   "GET",
		"headers":  map[string]string{"Authorization": "Bearer token"},
		"body":     "",
		"timeout":  10,
	}
	payloadBytes, _ := json.Marshal(payload)
	
	reqBody := service.CreateTaskRequest{
		Type:     models.TaskTypeHTTPRequest,
		Payload:  payloadBytes,
		Priority: 5,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var task models.Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	assert.NoError(t, err)
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, models.TaskTypeHTTPRequest, task.Type)
	assert.Equal(t, 5, task.Priority)
	assert.Equal(t, models.TaskStatePending, task.State)
	assert.JSONEq(t, string(payloadBytes), string(task.Payload))
}

func TestCreateTask_InvalidType(t *testing.T) {
	router, _ := setupTestRouter(t)

	reqBody := map[string]interface{}{
		"type":     "invalid_type",
		"payload":  map[string]any{"data": "test"},
		"priority": 5,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetTask(t *testing.T) {
	router, repository := setupTestRouter(t)

	task := models.NewTask(
		models.TaskTypeHTTPRequest,
		json.RawMessage(`{"url": "https://example.com"}`),
		5,
	)
	testutil.CreateTestTask(t, repository.DB(), task)

	req, _ := http.NewRequest("GET", "/api/v1/tasks/"+task.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var fetchedTask models.Task
	err := json.Unmarshal(w.Body.Bytes(), &fetchedTask)
	assert.NoError(t, err)
	assert.Equal(t, task.ID, fetchedTask.ID)
	assert.Equal(t, task.Type, fetchedTask.Type)
	assert.Equal(t, task.Priority, fetchedTask.Priority)
	assert.Equal(t, task.State, fetchedTask.State)
	assert.JSONEq(t, string(task.Payload), string(fetchedTask.Payload))
}

func TestGetTask_NotFound(t *testing.T) {
	router, _ := setupTestRouter(t)
	
	req, _ := http.NewRequest("GET", "/api/v1/tasks/nonexistent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListTasks(t *testing.T) {
	router, repository := setupTestRouter(t)

	// Create test tasks
	for i := 1; i <= 3; i++ {
		task := models.NewTask(
			models.TaskTypeHTTPRequest,
			json.RawMessage(`{"url": "https://example.com"}`),
			i,
		)
		testutil.CreateTestTask(t, repository.DB(), task)
	}

	req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)

	tasks := res["tasks"].([]any)
	assert.Len(t, tasks, 3)
	assert.Equal(t, float64(3), res["count"].(float64))
}

func TestListTasks_WithFilters(t *testing.T) {
	router, repository := setupTestRouter(t)

	pendiungTask := models.NewTask(
		models.TaskTypeHTTPRequest,
		json.RawMessage(`{"url": "https://example.com/pending"}`),
		5,
	)
	testutil.CreateTestTask(t, repository.DB(), pendiungTask)

	completedTask := models.NewTask(
		models.TaskTypeHTTPRequest,
		json.RawMessage(`{"url": "https://example.com/completed"}`),
		5,
	)
	completedTask.State = models.TaskStateCompleted
	testutil.CreateTestTask(t, repository.DB(), completedTask)

	// list only pending tasks
	req, _ := http.NewRequest("GET", "/api/v1/tasks?state=pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var res map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)

	tasks := res["tasks"].([]any)
	assert.Len(t, tasks, 1)
	assert.Equal(t, float64(1), res["count"].(float64))
}

func TestCancelTask(t *testing.T) {
	router, repository := setupTestRouter(t)

	task := models.NewTask(
		models.TaskTypeHTTPRequest,
		json.RawMessage(`{"url": "https://example.com"}`),
		5,
	)
	testutil.CreateTestTask(t, repository.DB(), task)

	req, _ := http.NewRequest("DELETE", "/api/v1/tasks/"+task.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	cancelled := testutil.GetTestTaskByID(t, repository.DB(), task.ID)
	assert.Equal(t, models.TaskStateCancelled, cancelled.State)
}

func TestCancelTask_AlreadyCompleted(t *testing.T) {
	router, repository := setupTestRouter(t)

	task := models.NewTask(
		models.TaskTypeHTTPRequest,
		json.RawMessage(`{"url": "https://example.com"}`),
		5,
	)
	task.State = models.TaskStateCompleted
	testutil.CreateTestTask(t, repository.DB(), task)

	req, _ := http.NewRequest("DELETE", "/api/v1/tasks/"+task.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	completed := testutil.GetTestTaskByID(t, repository.DB(), task.ID)
	assert.Equal(t, models.TaskStateCompleted, completed.State)
}
