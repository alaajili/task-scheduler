-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    priority INTEGER NOT NULL DEFAULT 5 CHECK (priority >= 0 AND priority <= 10),
    state VARCHAR(20) NOT NULL DEFAULT 'pending',
    result JSONB,
    error TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    worker_id VARCHAR(36),
    CONSTRAINT valid_state CHECK (state IN ('pending', 'running', 'completed', 'failed', 'cancelled'))
);

-- Create indexes for common queries
CREATE INDEX idx_tasks_state ON tasks(state);
CREATE INDEX idx_tasks_priority ON tasks(priority DESC);
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);
CREATE INDEX idx_tasks_type ON tasks(type);
CREATE INDEX idx_tasks_worker_id ON tasks(worker_id);

-- Composite index for queue polling (most important)
CREATE INDEX idx_tasks_queue ON tasks(state, priority DESC, created_at ASC)
WHERE state = 'pending';

-- Create workers table
CREATE TABLE IF NOT EXISTS workers (
    id VARCHAR(36) PRIMARY KEY,
    status VARCHAR(20) NOT NULL DEFAULT 'idle',
    current_task_id VARCHAR(36),
    last_heartbeat TIMESTAMP NOT NULL DEFAULT NOW(),
    task_types TEXT[] NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_status CHECK (status IN ('active', 'idle', 'shutdown')),
    CONSTRAINT fk_current_task FOREIGN KEY (current_task_id) REFERENCES tasks(id)
);

-- Create index for heartbeat checks
CREATE INDEX idx_workers_last_heartbeat ON workers(last_heartbeat DESC);
CREATE INDEX idx_workers_status ON workers(status);
