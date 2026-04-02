package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/cron"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/handlers"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/services"
	"github.com/user/remna-user-panel/internal/telegram"
)

func main() {
	log.Println("[main] starting Remna User Panel Backend...")

	// Load config
	configPath := "config.json"
	if env := os.Getenv("CONFIG_PATH"); env != "" {
		configPath = env
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		// If config doesn't exist, copy from example
		if os.IsNotExist(err) {
			log.Println("[main] config.json not found, creating from example...")
			exampleData, readErr := os.ReadFile("config.example.json")
			if readErr != nil {
				log.Fatalf("[main] failed to read config.example.json: %v", readErr)
			}
			if writeErr := os.WriteFile(configPath, exampleData, 0644); writeErr != nil {
				log.Fatalf("[main] failed to create config.json: %v", writeErr)
			}
			cfg, err = config.Load(configPath)
			if err != nil {
				log.Fatalf("[main] failed to load config: %v", err)
			}
		} else {
			log.Fatalf("[main] failed to load config: %v", err)
		}
	}
	config.WatchConfig()

	// Initialize database
	dbPath := "data/panel.db"
	if env := os.Getenv("DB_PATH"); env != "" {
		dbPath = env
	}
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("[main] database init: %v", err)
	}
	defer database.Close()

	// Initialize services
	creditSvc := services.NewCreditService()

	// Initialize handler
	h := handlers.NewHandler()

	// Initialize router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(chimw.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Telegram-Init-Data"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		middleware.WriteSuccess(w, map[string]string{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public payment callbacks (no auth)
		r.Post("/payment/callback/bepusdt", h.BEPusdtCallback)
		r.Post("/payment/callback/ezpay", h.EZPayCallback)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.CombinedAuth)

			// User
			r.Get("/user/me", h.GetMe)

			// Credit
			r.Get("/credit/balance", h.GetCreditBalance)
			r.Post("/credit/signup", h.CreditSignup)
			r.Post("/credit/bet", h.CreditBet)
			r.Get("/credit/history", h.GetCreditHistory)

			// Combos
			r.Get("/combos", h.ListCombos)

			// Subscription
			r.Post("/subscribe", h.PurchaseCombo)
			r.Post("/bind-sub", h.BindSubscription)
			r.Get("/sub-info", h.GetSubInfo)
			r.Get("/sub-keys", h.GetSubKeys)

			// Payment
			r.Post("/payment/create", h.CreatePayment)
			r.Post("/payment/custom", h.CustomPayment)
			r.Get("/orders", h.ListOrders)
			r.Get("/orders/{uuid}", h.GetOrder)

			// VPN Info
			r.Get("/vpn/bandwidth", h.GetBandwidthStats)
			r.Get("/vpn/devices", h.GetHWIDDevices)
			r.Get("/vpn/ips", h.GetIPList)
			r.Get("/vpn/history", h.GetSubHistory)

			// Squads
			r.Get("/squads/external", h.GetExternalSquads)
			r.Put("/squads/external", h.UpdateExternalSquad)

			// IP Change
			r.Post("/ip/change", h.IPChange)
			r.Get("/ip/status", h.GetIPChangeStatus)

			// Jellyfin
			r.Post("/jellyfin/purchase", h.PurchaseJellyfin)
			r.Post("/jellyfin/quick-connect", h.JellyfinQuickConnect)
			r.Put("/jellyfin/password", h.JellyfinUpdatePassword)
			r.Get("/jellyfin/devices", h.JellyfinGetDevices)
			r.Put("/jellyfin/parental-rating", h.JellyfinUpdateParentalRating)

			// Admin routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminOnly)

				r.Get("/admin/config", h.GetConfig)
				r.Put("/admin/config", h.UpdateConfig)
				r.Get("/admin/combos", h.AdminListCombos)
				r.Post("/admin/combos", h.CreateCombo)
				r.Put("/admin/combos/{uuid}", h.UpdateCombo)
				r.Delete("/admin/combos/{uuid}", h.DeleteCombo)
				r.Get("/admin/squads/internal", h.GetInternalSquads)
				r.Get("/admin/users", h.AdminListUsers)
				r.Get("/admin/users/{id}", h.AdminGetUser)
				r.Put("/admin/users/{id}", h.AdminUpdateUser)
				r.Get("/admin/orders", h.AdminListOrders)
				r.Put("/admin/orders/{uuid}", h.AdminUpdateOrder)
				r.Post("/admin/orders/{uuid}/actions/{action}", h.AdminOrderAction)
			})
		})
	})

	// Start cron jobs
	cron.Start(creditSvc, h.Payment)

	// Start Telegram bot
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot, err := telegram.NewBot(creditSvc)
	if err != nil {
		log.Printf("[main] telegram bot init: %v", err)
	}
	if bot != nil {
		go bot.Start(ctx)
	}

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("[main] HTTP server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[main] shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)

	log.Println("[main] server stopped")
}
