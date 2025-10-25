# Distributed Task Scheduler

A production-grade distributed job queue system with auto-scaling capabilities.

## Features
- REST API for task submission
- Priority-based task queuing
- Distributed worker pool with auto-scaling
- Leader election and distributed locking
- Complete observability (metrics, traces, logs)

## Tech Stack
- Go 1.25.1
- PostgreSQLclear
- Redis
- etcd
- Kubernetes
- Prometheus & Grafana

## Getting Started

### Prerequisites
- Go 1.25.1
- Docker & Docker Compose
- Make

### Local Development
```bash
make setup
make dev
```

## Project Structure
See [docs/architecture.md](docs/architecture.md)

## License
MIT
