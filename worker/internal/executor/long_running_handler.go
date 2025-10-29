package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alaajili/task-scheduler/shared/logger"
	"go.uber.org/zap"
)

// LongRunningPayload represents the payload for long-running tasks
type LongRunningPayload struct {
	DurationSeconds int    `json:"duration_seconds"`
	StepCount       int    `json:"step_count"`
	SimulateError   bool   `json:"simulate_error"`
	ErrorAfter      int    `json:"error_after"` // seconds
}

// LongRunningResult represents the result of a long-running task
type LongRunningResult struct {
	Completed     bool      `json:"completed"`
	Duration      float64   `json:"duration_seconds"`
	StepsExecuted int       `json:"steps_executed"`
	CompletedAt   time.Time `json:"completed_at"`
}

func (e *Executor) executeLongRunning(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var req LongRunningPayload
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("invalid long-running payload: %w", err)
	}

	// Default values
	if req.DurationSeconds <= 0 {
		req.DurationSeconds = 10
	}
	if req.StepCount <= 0 {
		req.StepCount = 5
	}

	logger.Info("Executing long-running task",
		zap.Int("duration_seconds", req.DurationSeconds),
		zap.Int("steps", req.StepCount),
	)

	startTime := time.Now()
	stepDuration := time.Duration(req.DurationSeconds) * time.Second / time.Duration(req.StepCount)

	// Execute steps
	for i := 0; i < req.StepCount; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("task cancelled: %w", ctx.Err())
		default:
			// Simulate error if requested
			if req.SimulateError && int(time.Since(startTime).Seconds()) >= req.ErrorAfter {
				return nil, fmt.Errorf("simulated error at step %d", i+1)
			}

			logger.Debug("Executing step",
				zap.Int("step", i+1),
				zap.Int("total", req.StepCount),
			)

			time.Sleep(stepDuration)
		}
	}

	duration := time.Since(startTime)

	result := LongRunningResult{
		Completed:     true,
		Duration:      duration.Seconds(),
		StepsExecuted: req.StepCount,
		CompletedAt:   time.Now(),
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	logger.Info("Long-running task completed",
		zap.Float64("duration_seconds", result.Duration),
	)

	return resultBytes, nil
}
