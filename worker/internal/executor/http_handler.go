package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alaajili/task-scheduler/shared/logger"
	"go.uber.org/zap"
)

type HTTPRequestPayload struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    any               `json:"body"`
	Timeout int               `json:"timeout"` // in seconds
}

type HTTPRequestResult struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   float64		     `json:"duration_ms"`
}

func (e *Executor) executeHTTPRequest(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var reqPayload HTTPRequestPayload
	if err := json.Unmarshal(payload, &reqPayload); err != nil {
		return nil, err
	}

	// validte request fields
	if reqPayload.URL == "" {
		return nil, fmt.Errorf("invalid HTTP request payload: missing URL")
	}
	if reqPayload.Method == "" {
		reqPayload.Method = "GET"
	}

	// set timeout
	timeout := 30 * time.Second
	if reqPayload.Timeout > 0 {
		timeout = time.Duration(reqPayload.Timeout) * time.Second
	}

	// create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	var bodyReader io.Reader
	if reqPayload.Body != nil {
		bodyBytes, err := json.Marshal(reqPayload.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	httpReq, err := http.NewRequestWithContext(ctx, reqPayload.Method, reqPayload.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// set headers
	for key, value := range reqPayload.Headers {
		httpReq.Header.Set(key, value)
	}

	logger.Info("Executing HTTP request",
		zap.String("method", reqPayload.Method),
		zap.String("url", reqPayload.URL),
	)

	startTime := time.Now()
	httpResp, err := client.Do(httpReq)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body: %w", err)
	}

	// build the result
	result := HTTPRequestResult{
		StatusCode: httpResp.StatusCode,
		Headers:    make(map[string]string),
		Body:       string(respBody),
		Duration:   float64(duration.Milliseconds()),
	}

	for key, values := range httpResp.Header {
		if len(values) > 0 {
			result.Headers[key] = values[0]
		}
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HTTP request result: %w", err)
	}

	logger.Info("HTTP request completed",
		zap.Int("status_code", httpResp.StatusCode),
		zap.Float64("duration_ms", result.Duration),
	)

	return resultBytes, nil
}
