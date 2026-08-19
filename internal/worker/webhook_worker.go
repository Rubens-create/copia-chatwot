package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/queue"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/repository"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/webhook"
	"github.com/google/uuid"
)

type WebhookWorker struct {
	queue       *queue.RedisQueue
	webhookRepo repository.WebhookRepository
	dispatcher  webhook.WebhookDispatcher
	concurrency int
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func NewWebhookWorker(
	q *queue.RedisQueue,
	webhookRepo repository.WebhookRepository,
	dispatcher webhook.WebhookDispatcher,
	concurrency int,
) *WebhookWorker {
	if concurrency <= 0 {
		concurrency = 5
	}
	if dispatcher == nil {
		dispatcher = webhook.NewDispatcher()
	}
	return &WebhookWorker{
		queue:       q,
		webhookRepo: webhookRepo,
		dispatcher:  dispatcher,
		concurrency: concurrency,
		stopChan:    make(chan struct{}),
	}
}

func (w *WebhookWorker) Start(ctx context.Context) {
	log.Printf("[Worker] Starting Webhook Worker with concurrency %d", w.concurrency)

	// 1. Event fan-out workers (read from incoming_messages and create delivery jobs)
	for i := 0; i < 2; i++ {
		w.wg.Add(1)
		go w.eventFanoutLoop(ctx, i)
	}

	// 2. Webhook delivery workers (read from outgoing_webhooks and do HTTP dispatch)
	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		go w.deliveryLoop(ctx, i)
	}

	// 3. Retry scheduler ticker (polls ZSET for due retries)
	w.wg.Add(1)
	go w.retrySchedulerLoop(ctx)
}

func (w *WebhookWorker) Stop() {
	log.Printf("[Worker] Stopping Webhook Worker...")
	close(w.stopChan)
	w.wg.Wait()
	log.Printf("[Worker] Webhook Worker stopped gracefully")
}

func (w *WebhookWorker) eventFanoutLoop(ctx context.Context, id int) {
	defer w.wg.Done()
	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		default:
			data, err := w.queue.Dequeue(ctx, queue.QueueIncomingMessages, 2*time.Second)
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			if data == nil {
				continue
			}

			var job queue.EventJob
			if err := json.Unmarshal(data, &job); err != nil {
				log.Printf("[Worker] Invalid event job data: %v", err)
				continue
			}

			// Find matching active webhooks
			webhooks, err := w.webhookRepo.ListActiveWebhooks(ctx, job.AccountID, job.InboxID, job.Type)
			if err != nil {
				log.Printf("[Worker] Error fetching active webhooks: %v", err)
				continue
			}

			if len(webhooks) == 0 {
				continue
			}

			payloadBytes, err := json.Marshal(job.Event)
			if err != nil {
				log.Printf("[Worker] Error marshaling event payload: %v", err)
				continue
			}

			for _, wh := range webhooks {
				secret := ""
				if wh.Secret != nil {
					secret = *wh.Secret
				}
				deliveryJob := queue.WebhookDeliveryJob{
					ID:          uuid.New().String(),
					WebhookID:   wh.ID,
					URL:         wh.URL,
					Secret:      secret,
					Event:       job.Type,
					Payload:     payloadBytes,
					Attempt:     0,
					MaxAttempts: webhook.MaxRetryAttempts,
					CreatedAt:   time.Now().UTC(),
				}

				if err := w.queue.Enqueue(ctx, queue.QueueOutgoingWebhooks, deliveryJob); err != nil {
					log.Printf("[Worker] Error enqueuing delivery job for webhook %s: %v", wh.URL, err)
				}
			}
		}
	}
}

func (w *WebhookWorker) deliveryLoop(ctx context.Context, workerID int) {
	defer w.wg.Done()
	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		default:
			data, err := w.queue.Dequeue(ctx, queue.QueueOutgoingWebhooks, 2*time.Second)
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			if data == nil {
				continue
			}

			var job queue.WebhookDeliveryJob
			if err := json.Unmarshal(data, &job); err != nil {
				log.Printf("[Worker-%d] Invalid delivery job data: %v", workerID, err)
				continue
			}

			job.Attempt++
			now := time.Now().UTC()

			status, err := w.dispatcher.Dispatch(ctx, job.URL, job.Secret, job.Event, job.Payload)
			if err == nil {
				// Delivery Success
				log.Printf(`{"level":"info","event":"webhook_delivery_success","url":"%s","status":%d,"attempt":%d}`,
					job.URL, status, job.Attempt)

				_ = w.logDelivery(ctx, job, "success", status, "", nil)
				continue
			}

			// Delivery Failed
			errMsg := err.Error()
			log.Printf(`{"level":"warn","event":"webhook_delivery_failed","url":"%s","status":%d,"attempt":%d,"error":"%s"}`,
				job.URL, status, job.Attempt, errMsg)

			delay, canRetry := webhook.GetNextRetryDelay(job.Attempt)
			if canRetry {
				nextAttemptAt := now.Add(delay)
				_ = w.logDelivery(ctx, job, "pending", status, errMsg, &nextAttemptAt)

				log.Printf("[Worker-%d] Scheduling retry #%d for %s in %v", workerID, job.Attempt+1, job.URL, delay)
				if err := w.queue.ScheduleRetry(ctx, &job, delay); err != nil {
					log.Printf("[Worker-%d] Error scheduling retry: %v", workerID, err)
				}
			} else {
				// Max retries exceeded -> Dead-letter
				_ = w.logDelivery(ctx, job, "dead_letter", status, errMsg, nil)
				log.Printf(`{"level":"error","event":"webhook_dead_letter","url":"%s","attempts":%d,"error":"%s"}`,
					job.URL, job.Attempt, errMsg)

				if err := w.queue.SendToDeadLetter(ctx, &job); err != nil {
					log.Printf("[Worker-%d] Error sending to dead letter queue: %v", workerID, err)
				}
			}
		}
	}
}

func (w *WebhookWorker) retrySchedulerLoop(ctx context.Context) {
	defer w.wg.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			items, err := w.queue.FetchDueRetries(ctx)
			if err != nil {
				continue
			}
			for _, item := range items {
				var job queue.WebhookDeliveryJob
				if err := json.Unmarshal(item, &job); err == nil {
					_ = w.queue.Enqueue(ctx, queue.QueueOutgoingWebhooks, job)
				}
			}
		}
	}
}

func (w *WebhookWorker) logDelivery(ctx context.Context, job queue.WebhookDeliveryJob, status string, responseCode int, lastError string, nextAttemptAt *time.Time) error {
	now := time.Now().UTC()
	var errStr *string
	if lastError != "" {
		errStr = &lastError
	}
	var respCode *int
	if responseCode > 0 {
		respCode = &responseCode
	}

	attempt := &model.WebhookDeliveryAttempt{
		WebhookID:     job.WebhookID,
		Event:         job.Event,
		Payload:       job.Payload,
		URL:           job.URL,
		Attempts:      job.Attempt,
		Status:        status,
		LastError:     errStr,
		ResponseCode:  respCode,
		LastAttemptAt: &now,
		NextAttemptAt: nextAttemptAt,
	}

	_, err := w.webhookRepo.LogDeliveryAttempt(ctx, attempt)
	return err
}
