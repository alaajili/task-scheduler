package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/alaajili/task-scheduler/shared/config"
	"github.com/redis/go-redis/v9"
)


type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(cfg config.RedisConfig) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: 100,
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	
	return &RedisQueue{client: client}, nil
}

// publish the task to the redis queue with the priority as a score
func (q *RedisQueue) PublishTask(ctx context.Context, taskID string, priority int) error {
	err := q.client.ZAdd(ctx, "task_queue", redis.Z{
		Score:  float64(priority),
		Member: taskID,
	}).Err()
	
	if err != nil {
		return fmt.Errorf("failed to publish the task: %w", err)
	}
	
	return nil
}

// pop the task with the highest priority from the queue
func (q *RedisQueue) PopTask(ctx context.Context) (string, error) {
	res, err := q.client.ZPopMax(ctx, "task_queue", 1).Result()
	if err != nil {
		return "", fmt.Errorf("failed to pop the task: %w", err)
	}
	if len(res) == 0 {
		return "", nil // no task available
	}
	
	taskID := res[0].Member.(string)
	return taskID, nil
}

// get the number of tasks in the queue
func (q *RedisQueue) GetQueueDepth(ctx context.Context) (int64, error) {
	count, err := q.client.ZCard(ctx, "task_queue").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get the queue depth: %w", err)
	}
	
	return count, nil
}

// publish a task to be proccesed after a delay
func (q *RedisQueue) PublishDelayedTask(
	ctx context.Context,
	taskID string,
	priority int,
	delay time.Duration,
) error {
	executeAt := time.Now().UTC().Add(delay).Unix()
	
	err := q.client.ZAdd(ctx, "delayed_queue", redis.Z{
		Score:  float64(executeAt),
		Member: fmt.Sprintf("%s:%d", taskID, priority),
	}).Err()
	
	if err != nil {
		return fmt.Errorf("failed to publish delayed task: %w", err)
	}
	
	return nil
}

// get the ready delayed tasks to be executed
func (q *RedisQueue) GetReadyDelayedTasks(ctx context.Context) ([]string, error) {
	now := time.Now().UTC().Unix()
	
	res, err := q.client.ZRangeByScore(ctx, "delayed_queue", &redis.ZRangeBy{
		Min:  "0",
		Max:  fmt.Sprintf("%d", now),
		Offset: 0,
		Count:  100,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get the ready delayed tasks: %w", err)
	}
	
	return res, nil
}

// remove a task from the delayed queue
func (q *RedisQueue) RemoveDelayedTask(ctx context.Context, taskWithPriority string) error {
	err := q.client.ZRem(ctx, "delayed_queue", taskWithPriority).Err()
	if err != nil {
		return fmt.Errorf("failed to remove delayed task: %w", err)
	}
	return nil
}

func (q *RedisQueue) Close() error {
	return q.client.Close()
}

func (q *RedisQueue) HealthCheck(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}
