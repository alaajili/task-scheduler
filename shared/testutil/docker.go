package testutil

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func EnsureDockerRunning(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker-compose", "ps", "-q")
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		t.Log("Starting Docker containers for tests...")

		cmd = exec.CommandContext(ctx, "docker-compose", "up", "-d")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to start Docker containers: %v", err)
		}

		// Wait a bit for the services to be ready
		time.Sleep(10 * time.Second)
	}
}

func StopDocker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker-compose", "down")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to stop Docker containers: %v", err)
	}
}
