# Distributed Task Scheduler

A production-grade distributed job queue system with auto-scaling capabilities.

## Current Status: Api Sever completed

- ✅ Project structure setup
- ✅ Core data models
- ✅ Database schema and migrations
- ✅ API Server implementation
- ✅ Configuration management
- ✅ Logging infrastructure
- ✅ Docker Compose for local development
- ✅ Integration tests

## Features (api-server)

- REST API for task submission
- Task state management (pending, running, completed, failed, cancelled)
- Priority-based queuing (0-10)
- PostgreSQL storage with optimized indexes
- Structured logging with zap
- Health and readiness endpoints
- Comprehensive test coverage

## Quick Start

### Prerequisites
- Go 1.25
- Docker & Docker Compose
- Make
- PostgreSQL client
- Redis client

### Setup
```bash
# Clone the repository
git clone https://www.github.com/alaajili/task-scheduler
cd task-scheduler

# Run setup script
./scripts/setup.sh

# Or manually:
make setup
make dev
```

### Running the API Server
```bash
# Start infrastructure
make docker-up

# Run migrations
make migrate-up

# Start API server
make run-api
```

### Testing
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests only
cd api-server && go test -v -tags=integration ./...
```

### Example API Usage

**Create a task:**
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "http_request",
    "payload": {"url": "https://example.com"},
    "priority": 5,
    "max_retries": 3
  }'
```

**Get task status:**
```bash
curl http://localhost:8080/api/v1/tasks/{task-id}
```

**List tasks:**
```bash
curl http://localhost:8080/api/v1/tasks?state=pending&limit=10
```

**Cancel a task:**
```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/{task-id}
```

## Project Structure
```
task-scheduler/
├── api-server/          # REST API service
│   ├── cmd/            # Main application
│   └── internal/       # Internal packages
│       ├── handlers/   # HTTP handlers
│       ├── repository/ # Data access layer
│       └── service/    # Business logic
├── scheduler/          # Task scheduler
├── worker/            # Task worker
├── shared/            # Shared libraries
│   ├── models/        # Data models
│   ├── database/      # Database utilities
│   ├── config/        # Configuration
│   └── logger/        # Logging
├── deploy/            # Deployment configs
│   ├── docker/        # Dockerfiles
│   └── kubernetes/    # K8s manifests
├── migrations/        # Database migrations
├── scripts/           # Utility scripts
└── tests/            # Integration tests
```

## Tech Stack

- **Language**: Go 1.25
- **Database**: PostgreSQL 15
- **Queue**: Redis 7
- **Coordination**: etcd 3.5
- **Framework**: Gin (HTTP)
- **Logging**: Zap
- **Migrations**: golang-migrate

## Configuration

Configuration can be set via `config.yaml` or environment variables:
```yaml
server:
  port: 8080

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: taskscheduler

redis:
  host: localhost
  port: 6379

etcd:
  endpoints:
    - localhost:2379
```

## Development

### Make Commands
```bash
make help              # Show all commands
make setup             # Install dependencies
make dev               # Start dev environment
make docker-up         # Start Docker services
make docker-down       # Stop Docker services
make migrate-up        # Run migrations
make migrate-down      # Rollback migrations
make test              # Run tests
make lint              # Run linters
make build-all         # Build all services
```

### Adding a New Migration
```bash
make migrate-create NAME=add_new_field
```

### Running Linters
```bash
make lint
```

## Testing

The project has comprehensive test coverage:

- **Unit tests**: Test individual functions and methods
- **Integration tests**: Test API endpoints with real database
- **Repository tests**: Test data access layer

Run tests with:
```bash
make test
```

View coverage:
```bash
make test-coverage
open shared/coverage.html
```

## Next:

### Worker Implementation
- [ ] Worker service with task execution
- [ ] Support for multiple task types
- [ ] Retry logic with exponential backoff
- [ ] Heartbeat mechanism

### Distributed Coordination
- [ ] etcd integration for service discovery
- [ ] Leader election for scheduler
- [ ] Distributed locking
- [ ] Redis queue integration

### Kubernetes Deployment
- [ ] Dockerfiles for all services
- [ ] Kubernetes manifests
- [ ] Horizontal Pod Autoscaler
- [ ] Helm charts

### Observability
- [ ] Prometheus metrics
- [ ] Distributed tracing (Jaeger)
- [ ] Grafana dashboards
- [ ] Structured logging aggregation

### Advanced Features
- [ ] Task dependencies (DAG)
- [ ] Cron/scheduled tasks
- [ ] Rate limiting per task type
- [ ] Admin dashboard

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make test` and `make lint`
6. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details

## Support

For questions or issues:
- Open an issue on GitHub
- Check the [docs](docs/) folder for detailed documentation

---

**Status**: Complete - API Server fully functional

**Next**: Worker Implementation