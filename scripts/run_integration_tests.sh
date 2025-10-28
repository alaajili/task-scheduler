#!/bin/bash

set -e

echo "🧪 Running integration tests..."

# Start services if not running
if ! docker-compose ps | grep -q "Up"; then
    echo "Starting Docker services..."
    make docker-up
    sleep 10
fi

# Run migrations
echo "Running migrations..."
make migrate-up

# Start API server in background
echo "Starting API server..."
cd api-server
go run cmd/main.go &
API_PID=$!
cd ..

# Wait for API to be ready
echo "Waiting for API server to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:8080/health > /dev/null; then
        echo "API server is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "API server failed to start"
        kill $API_PID
        exit 1
    fi
    sleep 1
done

# Run integration tests
echo "Running integration tests..."
cd tests/integration
go test -v -tags=integration ./...
TEST_RESULT=$?

# Cleanup
echo "Cleaning up..."
kill $API_PID

if [ $TEST_RESULT -eq 0 ]; then
    echo "✅ All integration tests passed!"
else
    echo "❌ Some tests failed"
    exit 1
fi
