package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/config"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/database"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/queue"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/repository"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/webhook"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/worker"
)

func main() {
	log.Println("[Worker] Starting Chatwoot-Lite Async Worker...")

	cfg := config.Load()

	// 1. Connect to PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[Worker] Database connection error: %v", err)
	}
	defer db.Close()

	// 2. Connect to Redis
	redisQueue, err := queue.ConnectRedis(cfg.RedisURL)
	if err != nil {
		log.Fatalf("[Worker] Redis connection error: %v", err)
	}
	defer redisQueue.Close()

	// 3. Initialize Repositories and Dispatcher
	webhookRepo := repository.NewWebhookRepository(db)
	dispatcher := webhook.NewDispatcher()

	// 4. Initialize Worker
	w := worker.NewWebhookWorker(redisQueue, webhookRepo, dispatcher, cfg.WorkerConcurrency)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	w.Start(workerCtx)

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("[Worker] Shutting down worker...")

	workerCancel()
	w.Stop()
	log.Println("[Worker] Worker exiting")
}
