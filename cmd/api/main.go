package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/config"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/database"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/handler"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/queue"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/repository"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/service"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/whatsapp"
)

func main() {
	log.Println("[API] Starting Chatwoot-Lite WhatsApp Gateway API...")

	cfg := config.Load()

	// 1. Connect to PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[API] Database connection error: %v", err)
	}
	defer db.Close()

	// Auto-run schema migration if migration file exists
	migrationFiles := []string{
		"migrations/001_chatwoot_whatsapp_schema.sql",
		"/app/migrations/001_chatwoot_whatsapp_schema.sql",
	}
	for _, mFile := range migrationFiles {
		if content, err := os.ReadFile(mFile); err == nil {
			if err := db.RunMigration(context.Background(), string(content)); err != nil {
				log.Printf("[API] Warning: Migration execution returned: %v", err)
			}
			break
		}
	}

	// 2. Connect to Redis
	redisQueue, err := queue.ConnectRedis(cfg.RedisURL)
	if err != nil {
		log.Fatalf("[API] Redis connection error: %v", err)
	}
	defer redisQueue.Close()

	// 3. Initialize Repositories
	accountRepo := repository.NewAccountRepository(db)
	contactRepo := repository.NewContactRepository(db)
	convRepo := repository.NewConversationRepository(db)
	msgRepo := repository.NewMessageRepository(db)
	eventRepo := repository.NewEventRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)

	// 4. Initialize Services
	waClient := whatsapp.NewClient(cfg.MetaAccessToken, cfg.MetaAPIVersion)
	waService := service.NewWhatsAppService(cfg, db, accountRepo, contactRepo, convRepo, msgRepo, eventRepo, redisQueue)
	convService := service.NewConversationService(cfg, accountRepo, contactRepo, convRepo, msgRepo, waClient, redisQueue)

	// 5. Initialize Handlers
	healthHandler := handler.NewHealthHandler(db, redisQueue)
	metaHandler := handler.NewMetaWebhookHandler(cfg, waService)
	convHandler := handler.NewConversationHandler(convService)
	webhookHandler := handler.NewWebhookHandler(cfg, webhookRepo)

	// 6. Router Setup
	mux := http.NewServeMux()

	// Health Checks
	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/ready", healthHandler.Ready)

	// Meta WhatsApp Webhook Endpoint & Aliases
	mux.HandleFunc("/webhooks/whatsapp", metaHandler.HandleWebhook)
	mux.HandleFunc("/webhook/whatsapp", metaHandler.HandleWebhook)
	mux.HandleFunc("/webhook", metaHandler.HandleWebhook)
	mux.HandleFunc("/webhooks/meta", metaHandler.HandleWebhook)

	// Meta Configuration API (GET and POST for UI Settings & Webhook Callback info)
	metaConfigHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		proto := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			proto = "https"
		}

		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPost {
			var payload struct {
				PhoneNumberID string `json:"phone_number_id"`
				PhoneNumber   string `json:"phone_number"`
				AccessToken   string `json:"access_token"`
				AppSecret     string `json:"app_secret"`
				VerifyToken   string `json:"verify_token"`
				APIVersion    string `json:"api_version"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, `{"error":"payload inválido"}`, http.StatusBadRequest)
				return
			}

			// Update in PostgreSQL channel_whatsapp table
			if err := accountRepo.UpdateChannelWhatsAppConfig(r.Context(), cfg.DefaultAccountID, payload.PhoneNumber, payload.PhoneNumberID, payload.AccessToken, payload.APIVersion); err != nil {
				log.Printf("[API] Error updating channel_whatsapp config: %v", err)
			}

			// Update in-memory runtime configurations
			if payload.PhoneNumberID != "" {
				cfg.MetaPhoneNumberID = payload.PhoneNumberID
			}
			if payload.AccessToken != "" {
				cfg.MetaAccessToken = payload.AccessToken
			}
			if payload.AppSecret != "" {
				cfg.MetaAppSecret = payload.AppSecret
			}
			if payload.VerifyToken != "" {
				cfg.MetaVerifyToken = payload.VerifyToken
			}
			if payload.APIVersion != "" {
				cfg.MetaAPIVersion = payload.APIVersion
			}
			waClient.UpdateCredentials(cfg.MetaAccessToken, cfg.MetaAPIVersion)

			log.Printf("[API] Meta WhatsApp credentials updated via Settings UI. PhoneID=%s, Version=%s", cfg.MetaPhoneNumberID, cfg.MetaAPIVersion)

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "ok",
				"message": "Credenciais da Meta salvas com sucesso!",
			})
			return
		}

		// GET request: Return active Meta configs
		phoneID := cfg.MetaPhoneNumberID
		phoneNumber := ""
		accessToken := cfg.MetaAccessToken

		if ch, err := accountRepo.GetDefaultChannelWhatsApp(r.Context(), cfg.DefaultAccountID); err == nil && ch != nil {
			if phoneNumber == "" {
				phoneNumber = ch.PhoneNumber
			}
			if accessToken == "" && ch.BusinessManagementToken != nil {
				accessToken = *ch.BusinessManagementToken
			}
			if phoneID == "" && len(ch.ProviderConfig) > 0 {
				var pConfig map[string]interface{}
				_ = json.Unmarshal(ch.ProviderConfig, &pConfig)
				if pid, ok := pConfig["phone_number_id"].(string); ok && pid != "" {
					phoneID = pid
				}
			}
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"webhook_url":      fmt.Sprintf("%s://%s/webhooks/whatsapp", proto, host),
			"webhook_path":     "/webhooks/whatsapp",
			"verify_token":     cfg.MetaVerifyToken,
			"api_access_token": cfg.APIAccessToken,
			"required_fields":  []string{"messages"},
			"api_version":      cfg.MetaAPIVersion,
			"phone_number_id":  phoneID,
			"phone_number":     phoneNumber,
			"access_token":     accessToken,
			"app_secret":       cfg.MetaAppSecret,
		})
	})
	mux.Handle("/api/config/meta", metaConfigHandler)
	mux.Handle("/api/config/meta-webhook", metaConfigHandler)

	// Auth Middleware for protected API endpoints
	authWrapper := handler.AuthMiddleware(cfg.APIAccessToken)

	// Protected Conversation Routes (Chatwoot v1 & Legacy Aliases)
	convRouter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/messages") {
			convHandler.HandleMessages(w, r)
		} else if strings.HasSuffix(path, "/conversations") || strings.HasSuffix(path, "/conversations/") {
			convHandler.ListConversations(w, r)
		} else {
			convHandler.GetConversation(w, r)
		}
	})

	// 1. Chatwoot Official v1 Endpoints:
	// GET  /api/v1/accounts/{account_id}/conversations
	// GET  /api/v1/accounts/{account_id}/conversations/{conversation_id}
	// GET  /api/v1/accounts/{account_id}/conversations/{conversation_id}/messages
	// POST /api/v1/accounts/{account_id}/conversations/{conversation_id}/messages
	mux.Handle("/api/v1/accounts/", authWrapper(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/webhooks") {
			webhookHandler.HandleWebhooks(w, r)
		} else {
			convRouter.ServeHTTP(w, r)
		}
	})))

	// 2. Legacy / Direct API Endpoints:
	mux.Handle("/api/conversations", authWrapper(convRouter))
	mux.Handle("/api/conversations/", authWrapper(convRouter))
	mux.Handle("/api/webhooks", authWrapper(http.HandlerFunc(webhookHandler.HandleWebhooks)))

	// Static Web UI Files
	staticDirs := []string{"web/static", "/app/web/static"}
	var staticDir string
	for _, dir := range staticDirs {
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			staticDir = dir
			break
		}
	}
	if staticDir != "" {
		fileServer := http.FileServer(http.Dir(staticDir))
		mux.Handle("/", fileServer)
	}

	// Apply Middlewares
	handlerWithMiddleware := handler.LoggerMiddleware(handler.CORSMiddleware(handler.RecoveryMiddleware(mux)))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.AppPort),
		Handler:      handlerWithMiddleware,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful Shutdown Channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[API] Server listening on port :%s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[API] Server failed: %v", err)
		}
	}()

	<-stop
	log.Println("[API] Shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[API] Server shutdown error: %v", err)
	}
	log.Println("[API] Server exiting")
}
