package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gogame.uz/backend/internal/config"
	"gogame.uz/backend/internal/database"
	"gogame.uz/backend/internal/routes"
	ws "gogame.uz/backend/internal/websocket"
)

func main() {
	// Load config
	config.Load()

	// Connect DB
	database.Connect()
	defer database.Close()

	// Run migrations
	database.Migrate()

	// Create WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Cognita.uz",
		ErrorHandler: errorHandler,
		BodyLimit:    int(config.App.MaxFileSize),
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.App.FrontendURL,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Routes
	routes.Setup(app, database.DB, hub)

	// Static files (uploads)
	app.Static("/uploads", config.App.UploadDir)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := ":" + config.App.Port
		log.Printf("Cognita.uz server starting on %s (env=%s)", addr, config.App.Env)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
