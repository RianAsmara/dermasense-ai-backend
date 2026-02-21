package http

import (
	"time"

	"github.com/dermasense/backend/internal/delivery/http/handler"
	"github.com/dermasense/backend/internal/delivery/http/middleware"
	"github.com/dermasense/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"
)

// Router configures all HTTP routes.
type Router struct {
	app         *fiber.App
	authHandler *handler.AuthHandler
	scanHandler *handler.ScanHandler
	authService *service.AuthService
	redisClient *redis.Client
}

// NewRouter creates a new Router.
func NewRouter(
	app *fiber.App,
	authHandler *handler.AuthHandler,
	scanHandler *handler.ScanHandler,
	authService *service.AuthService,
	redisClient *redis.Client,
) *Router {
	return &Router{
		app:         app,
		authHandler: authHandler,
		scanHandler: scanHandler,
		authService: authService,
		redisClient: redisClient,
	}
}

// Setup registers all routes and middleware.
func (r *Router) Setup() {
	r.app.Use(recover.New())
	r.app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))
	r.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	r.app.Static("/uploads", "./uploads")

	r.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "dermasense-api",
		})
	})

	v1 := r.app.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.Use(middleware.RateLimiter(r.redisClient, middleware.RateLimiterConfig{
		Max:       10,
		Window:    1 * time.Minute,
		KeyPrefix: "rl:auth",
	}))
	auth.Post("/google", r.authHandler.GoogleLogin)

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(r.authService))

	scans := protected.Group("/scans")
	scans.Use(middleware.RateLimiter(r.redisClient, middleware.RateLimiterConfig{
		Max:       30,
		Window:    1 * time.Minute,
		KeyPrefix: "rl:scans",
	}))
	scans.Post("/", r.scanHandler.CreateScan)
	scans.Get("/", r.scanHandler.GetScans)
	scans.Get("/:id", r.scanHandler.GetScanByID)

	quota := protected.Group("/quota")
	quota.Get("/", r.scanHandler.GetQuota)
	quota.Post("/add", r.scanHandler.AddQuota)
}
