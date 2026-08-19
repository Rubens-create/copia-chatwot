package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	Client *redis.Client
}

func ConnectRedis(redisURL string) (*RedisQueue, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	client := redis.NewClient(opt)
	var pingErr error

	for attempt := 1; attempt <= 10; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		pingErr = client.Ping(ctx).Err()
		cancel()

		if pingErr == nil {
			log.Printf("[Redis] Connected successfully to Redis at %s", opt.Addr)
			return &RedisQueue{Client: client}, nil
		}

		log.Printf("[Redis] Connection attempt %d/10 failed: %v. Retrying in 2s...", attempt, pingErr)
		time.Sleep(2 * time.Second)
	}

	client.Close()
	return nil, fmt.Errorf("failed to connect to redis after 10 attempts: %w", pingErr)
}

func (q *RedisQueue) Close() error {
	if q.Client != nil {
		return q.Client.Close()
	}
	return nil
}

func (q *RedisQueue) Ping(ctx context.Context) error {
	if q.Client == nil {
		return fmt.Errorf("redis client is nil")
	}
	return q.Client.Ping(ctx).Err()
}

func (q *RedisQueue) Enqueue(ctx context.Context, queueName string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal queue item: %w", err)
	}

	return q.Client.LPush(ctx, queueName, bytes).Err()
}

func (q *RedisQueue) Dequeue(ctx context.Context, queueName string, timeout time.Duration) ([]byte, error) {
	res, err := q.Client.BRPop(ctx, timeout, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	if len(res) < 2 {
		return nil, nil
	}

	return []byte(res[1]), nil
}

func (q *RedisQueue) ScheduleRetry(ctx context.Context, job *WebhookDeliveryJob, delay time.Duration) error {
	bytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal retry job: %w", err)
	}

	readyAt := float64(time.Now().Add(delay).Unix())
	return q.Client.ZAdd(ctx, QueueRetry, redis.Z{
		Score:  readyAt,
		Member: string(bytes),
	}).Err()
}

func (q *RedisQueue) FetchDueRetries(ctx context.Context) ([][]byte, error) {
	now := float64(time.Now().Unix())
	opt := &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", now),
	}

	items, err := q.Client.ZRangeByScore(ctx, QueueRetry, opt).Result()
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil
	}

	results := make([][]byte, 0, len(items))
	for _, item := range items {
		removed, err := q.Client.ZRem(ctx, QueueRetry, item).Result()
		if err == nil && removed > 0 {
			results = append(results, []byte(item))
		}
	}

	return results, nil
}

func (q *RedisQueue) SendToDeadLetter(ctx context.Context, job *WebhookDeliveryJob) error {
	bytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal dead letter job: %w", err)
	}

	return q.Client.LPush(ctx, QueueDeadLetter, bytes).Err()
}
