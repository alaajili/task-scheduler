package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alaajili/task-scheduler/shared/logger"
	"go.uber.org/zap"
)

// EmailPayload represents the payload for email tasks
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	From    string `json:"from"`
}

// EmailResult represents the result of sending an email
type EmailResult struct {
	Sent      bool      `json:"sent"`
	MessageID string    `json:"message_id"`
	SentAt    time.Time `json:"sent_at"`
}

func (e *Executor) executeEmailSend(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var req EmailPayload
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("invalid email payload: %w", err)
	}

	// Validate
	if req.To == "" {
		return nil, fmt.Errorf("recipient email is required")
	}
	if req.Subject == "" {
		return nil, fmt.Errorf("email subject is required")
	}

	logger.Info("Sending email",
		zap.String("to", req.To),
		zap.String("subject", req.Subject),
	)

	// Simulate email sending (in real implementation, use SMTP or email service)
	time.Sleep(100 * time.Millisecond)

	result := EmailResult{
		Sent:      true,
		MessageID: fmt.Sprintf("msg-%d", time.Now().Unix()),
		SentAt:    time.Now(),
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	logger.Info("Email sent successfully",
		zap.String("to", req.To),
		zap.String("message_id", result.MessageID),
	)

	return resultBytes, nil
}
