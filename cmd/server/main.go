package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dermasense/backend/internal/config"
	httpDelivery "github.com/dermasense/backend/internal/delivery/http"
	"github.com/dermasense/backend/internal/delivery/http/handler"
	"github.com/dermasense/backend/internal/repository/postgres"
	"github.com/dermasense/backend/internal/service"
	"github.com/dermasense/backend/pkg/database"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Load()

	pool, err := database.NewPostgresPool(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	log.Println("✓ Connected to PostgreSQL")

	if err := database.RunMigrations(pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	redisClient, err := database.NewRedisClient(cfg.RedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("✓ Connected to Redis")

	if err := os.MkdirAll("tmp/scans", 0o755); err != nil {
		log.Fatalf("Failed to create temp scan directory: %v", err)
	}

	userRepo := postgres.NewUserRepo(pool)
	scanRepo := postgres.NewScanRepo(pool)
	quotaRepo := postgres.NewQuotaRepo(pool)

	authService := service.NewAuthService(userRepo, quotaRepo, cfg)
	openAIService := service.NewOpenAIService(cfg)
	scanService := service.NewScanService(scanRepo, quotaRepo, openAIService, cfg)
	scanService.StartImageCleanup(3*24*time.Hour, 1*time.Hour) // Delete images older than 3 days, check every hour
	quotaService := service.NewQuotaService(quotaRepo)

	authHandler := handler.NewAuthHandler(authService)
	scanHandler := handler.NewScanHandler(scanService, quotaService)

	app := fiber.New(fiber.Config{
		AppName:               "DermaSense API",
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
		BodyLimit:             int(cfg.App.MaxFileSize) + 1024*1024,
		DisableStartupMessage: false,
	})

	router := httpDelivery.NewRouter(app, authHandler, scanHandler, authService, redisClient)
	router.Setup()

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		log.Printf("🚀 DermaSense API starting on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server exited gracefully")
}
