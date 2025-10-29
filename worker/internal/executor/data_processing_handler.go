package executor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alaajili/task-scheduler/shared/logger"
	"go.uber.org/zap"
)

type DataProcessingPayload struct {
	Operation string           `json:"operation"`
	Data      []map[string]any `json:"data"`
	Options   map[string]any   `json:"options"`
}

type DataProcessingResult struct {
	Operation    string `json:"operation"`
	Result       any    `json:"result"`
	RecordsCount int    `json:"records_count"`
}

func (e *Executor) executeDataProcessing(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var dpPayload DataProcessingPayload
	if err := json.Unmarshal(payload, &dpPayload); err != nil {
		return nil, fmt.Errorf("invalid data processing payload: %w", err)
	}

	logger.Info("Executing data processing task",
		zap.String("operation", dpPayload.Operation),
		zap.Int("records", len(dpPayload.Data)),
	)

	var result any
	var err error

	switch dpPayload.Operation {
	case "aggregate":
		result, err = e.aggregateData(dpPayload.Data, dpPayload.Options)
	case "filter":
		result, err = e.filterData(dpPayload.Data, dpPayload.Options)
	case "transform":
		result, err = e.transformData(dpPayload.Data, dpPayload.Options)
	default:
		return nil, fmt.Errorf("unsupported data processing operation: %s", dpPayload.Operation)
	}

	if err != nil {
		return nil, fmt.Errorf("data processing error: %w", err)
	}

	dpResult := DataProcessingResult{
		Operation:    dpPayload.Operation,
		Result:       result,
		RecordsCount: len(dpPayload.Data),
	}

	resultBytes, err := json.Marshal(dpResult)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data processing result: %w", err)
	}

	return resultBytes, nil
}

func (e *Executor) aggregateData(data []map[string]any, options map[string]any) (any, error) {
	logger.Info("Starting aggregation",
		zap.Int("records", len(data)),
		zap.Any("options", options),
	)

	// simple aggregation: count records
	result := map[string]any{
		"count": len(data),
		"type":  "aggregate",
	}

	logger.Info("Aggregation completed",
		zap.Any("result", result),
	)

	return result, nil
}

func (e *Executor) filterData(data []map[string]any, options map[string]any) (any, error) {
	logger.Info("Starting filtering",
		zap.Int("input_records", len(data)),
		zap.Any("options", options),
	)

	// simple filter: return records as is
	filtered := data

	logger.Info("Filtering completed",
		zap.Int("output_records", len(filtered)),
	)

	return filtered, nil
}

func (e *Executor) transformData(data []map[string]any, options map[string]any) (any, error) {
	logger.Info("Starting transformation",
		zap.Int("records", len(data)),
		zap.Any("options", options),
	)

	// simple transform: add processed=true to each record
	for i := range data {
		data[i]["processed"] = true
	}

	logger.Info("Transformation completed",
		zap.Int("records_processed", len(data)),
	)

	return data, nil
}
