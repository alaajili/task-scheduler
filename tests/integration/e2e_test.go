package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/alaajili/task-scheduler/shared/models"
)

const baseURL = "http://localhost:8080/api/v1"

func TestEndToEnd_TaskLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create a task
	taskReq := map[string]interface{}{
		"type":        "http_request",
		"payload":     map[string]interface{}{"url": "https://example.com"},
		"priority":    7,
		"max_retries": 3,
	}

	body, _ := json.Marshal(taskReq)
	resp, err := http.Post(baseURL+"/tasks", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var task models.Task
	json.NewDecoder(resp.Body).Decode(&task)
	resp.Body.Close()

	assert.NotEmpty(t, task.ID)
	assert.Equal(t, models.TaskTypeHTTPRequest, task.Type)
	assert.Equal(t, 7, task.Priority)
	assert.Equal(t, models.TaskStatePending, task.State)

	// Get the task
	resp, err = http.Get(fmt.Sprintf("%s/tasks/%s", baseURL, task.ID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var retrieved models.Task
	json.NewDecoder(resp.Body).Decode(&retrieved)
	resp.Body.Close()

	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, task.Type, retrieved.Type)

	// List tasks
	resp, err = http.Get(baseURL + "/tasks?state=pending")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResp)
	resp.Body.Close()

	tasks := listResp["tasks"].([]interface{})
	assert.GreaterOrEqual(t, len(tasks), 1)

	// Cancel the task
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/tasks/%s", baseURL, task.ID), nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify cancellation
	resp, err = http.Get(fmt.Sprintf("%s/tasks/%s", baseURL, task.ID))
	require.NoError(t, err)

	var cancelled models.Task
	json.NewDecoder(resp.Body).Decode(&cancelled)
	resp.Body.Close()

	assert.Equal(t, models.TaskStateCancelled, cancelled.State)
}

func TestHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	resp, err := http.Get("http://localhost:8080/health")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestReadinessCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	resp, err := http.Get("http://localhost:8080/ready")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
