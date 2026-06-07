package routes

import (
	"github.com/gofiber/fiber/v2"
	fws "github.com/gofiber/websocket/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"gogame.uz/backend/internal/handlers"
	"gogame.uz/backend/internal/middleware"
	"gogame.uz/backend/internal/models"
	ws "gogame.uz/backend/internal/websocket"
)

func Setup(app *fiber.App, db *pgxpool.Pool, hub *ws.Hub) {
	// ── Handlers ──────────────────────────────────────────────────────────────
	authH := handlers.NewAuthHandler(db)
	quizH := handlers.NewQuizHandler(db)
	roomH := handlers.NewRoomHandler(db, hub)
	classH := handlers.NewClassHandler(db)
	settingsH := handlers.NewSettingsHandler(db)
	aiH := handlers.NewAIHandler(db)

	// ── CORS & Global ─────────────────────────────────────────────────────────
	api := app.Group("/api")

	// ── Auth (public) ─────────────────────────────────────────────────────────
	auth := api.Group("/auth")
	auth.Post("/register", authH.Register)
	auth.Post("/login", authH.Login)
	auth.Post("/logout", authH.Logout)
	auth.Get("/me", middleware.Protected(), authH.Me)

	// ── Quizzes ───────────────────────────────────────────────────────────────
	// These must be registered before the group to avoid :id shadowing them
	api.Get("/quizzes/discover", middleware.Protected(), quizH.Discover)
	importH := handlers.NewImportHandler(aiH)
	api.Post("/quizzes/import-file", middleware.Protected(), middleware.RequireRole(models.RoleTeacher, models.RoleAdmin), importH.ImportFile)

	quizzes := api.Group("/quizzes", middleware.Protected())
	quizzes.Get("/", middleware.RequireRole(models.RoleTeacher, models.RoleAdmin), quizH.List)
	quizzes.Post("/", middleware.RequireRole(models.RoleTeacher, models.RoleAdmin), quizH.Create)
	quizzes.Get("/:id", quizH.Get)
	quizzes.Put("/:id", middleware.RequireRole(models.RoleTeacher, models.RoleAdmin), quizH.Update)
	quizzes.Delete("/:id", middleware.RequireRole(models.RoleTeacher, models.RoleAdmin), quizH.Delete)

	// ── Rooms ─────────────────────────────────────────────────────────────────
	rooms := api.Group("/rooms")
	rooms.Get("/history", middleware.Protected(), middleware.RequireRole(models.RoleTeacher, models.RoleAdmin), roomH.History)
	rooms.Post("/", middleware.Protected(), middleware.RequireRole(models.RoleTeacher, models.RoleAdmin), roomH.Create)
	rooms.Get("/:id/results", middleware.Protected(), middleware.RequireRole(models.RoleTeacher, models.RoleAdmin), roomH.Results)
	rooms.Get("/:pin/info", roomH.Info) // Public — for join screen

	// ── Classes ───────────────────────────────────────────────────────────────
	classGroup := api.Group("/classes", middleware.Protected(), middleware.RequireRole(models.RoleTeacher, models.RoleAdmin))
	classGroup.Get("/", classH.List)
	classGroup.Post("/", classH.Create)
	classGroup.Put("/:id", classH.Update)
	classGroup.Delete("/:id", classH.Delete)

	// ── Settings ──────────────────────────────────────────────────────────────
	api.Get("/settings", middleware.Protected(), settingsH.GetAll)
	api.Put("/settings", middleware.Protected(), middleware.RequireRole(models.RoleTeacher, models.RoleAdmin), settingsH.Set)

	// ── AI ────────────────────────────────────────────────────────────────────
	api.Post("/ai/generate-questions", middleware.Protected(), middleware.RequireRole(models.RoleTeacher, models.RoleAdmin), aiH.GenerateQuestions)
	api.Post("/ai/test", middleware.Protected(), middleware.RequireRole(models.RoleTeacher, models.RoleAdmin), aiH.TestConnection)


	// ── Exams & Olympiads ─────────────────────────────────────────────────────
	examH := handlers.NewExamHandler(db)
	olympiadH := handlers.NewOlympiadHandler(db)

	// Exams (public — students, registered BEFORE protected group to avoid shadowing)
	api.Get("/exams/join/:code", examH.JoinByCode)
	api.Get("/exams/take/:code", examH.TakeExam)
	api.Post("/exams/submit/:code", examH.Submit)

	// Olympiads (public — students, registered BEFORE protected group)
	api.Get("/olympiads/join/:code", olympiadH.JoinByCode)
	api.Get("/olympiads/take/:code", olympiadH.TakeOlympiad)
	api.Post("/olympiads/submit/:code", olympiadH.Submit)

	// Exams (teacher protected)
	exams := api.Group("/exams", middleware.Protected(), middleware.RequireRole(models.RoleTeacher, models.RoleAdmin))
	exams.Get("/", examH.List)
	exams.Post("/", examH.Create)
	exams.Get("/:id", examH.Get)
	exams.Put("/:id", examH.Update)
	exams.Put("/:id/status", examH.SetStatus)
	exams.Delete("/:id", examH.Delete)
	exams.Get("/:id/results", examH.Results)

	// Olympiads (teacher protected)
	olymps := api.Group("/olympiads", middleware.Protected(), middleware.RequireRole(models.RoleTeacher, models.RoleAdmin))
	olymps.Get("/", olympiadH.List)
	olymps.Post("/", olympiadH.Create)
	olymps.Get("/:id", olympiadH.Get)
	olymps.Put("/:id", olympiadH.Update)
	olymps.Put("/:id/status", olympiadH.SetStatus)
	olymps.Delete("/:id", olympiadH.Delete)
	olymps.Get("/:id/leaderboard", olympiadH.Leaderboard)

	// ── WebSocket Upgrade ─────────────────────────────────────────────────────
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// ws://server/ws/room/:pin?role=host&token=JWT  (host)
	// ws://server/ws/room/:pin?role=student         (student)
	app.Get("/ws/room/:pin", fws.New(roomH.WebSocketHandler))

	// ── Health ────────────────────────────────────────────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"ws":     hub.Stats(),
		})
	})
}
