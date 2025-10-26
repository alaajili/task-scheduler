DROP INDEX IF EXISTS idx_workers_status;
DROP INDEX IF EXISTS idx_workers_last_heartbeat;
DROP TABLE IF EXISTS workers;

DROP INDEX IF EXISTS idx_tasks_queue;
DROP INDEX IF EXISTS idx_tasks_worker_id;
DROP INDEX IF EXISTS idx_tasks_type;
DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_priority;
DROP INDEX IF EXISTS idx_tasks_state;
DROP TABLE IF EXISTS tasks;
