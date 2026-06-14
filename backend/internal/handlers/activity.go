package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gogame.uz/backend/internal/middleware"
)

// ActivityHandler — doska topshiriqlari (board_activities) CRUD.
type ActivityHandler struct {
	db *pgxpool.Pool
}

func NewActivityHandler(db *pgxpool.Pool) *ActivityHandler {
	return &ActivityHandler{db: db}
}

type activityDTO struct {
	ID        uuid.UUID       `json:"id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Content   json.RawMessage `json:"content,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type activityRequest struct {
	Type    string          `json:"type"`
	Title   string          `json:"title"`
	Content json.RawMessage `json:"content"`
}

// GET /api/activities — o'qituvchining topshiriqlari (content'siz, ro'yxat uchun)
func (h *ActivityHandler) List(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)
	rows, err := h.db.Query(context.Background(),
		`SELECT id, type, title, created_at, updated_at
		 FROM board_activities WHERE teacher_id = $1 ORDER BY updated_at DESC`,
		teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	list := make([]activityDTO, 0)
	for rows.Next() {
		var a activityDTO
		if err := rows.Scan(&a.ID, &a.Type, &a.Title, &a.CreatedAt, &a.UpdatedAt); err != nil {
			continue
		}
		list = append(list, a)
	}
	return c.JSON(list)
}

// GET /api/activities/:id — to'liq topshiriq (content bilan)
func (h *ActivityHandler) Get(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	var a activityDTO
	err = h.db.QueryRow(context.Background(),
		`SELECT id, type, title, content, created_at, updated_at
		 FROM board_activities WHERE id = $1 AND teacher_id = $2`,
		id, teacherID,
	).Scan(&a.ID, &a.Type, &a.Title, &a.Content, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Topshiriq topilmadi"})
	}
	return c.JSON(a)
}

// POST /api/activities
func (h *ActivityHandler) Create(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)
	var req activityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Title == "" || req.Type == "" {
		return c.Status(400).JSON(fiber.Map{"error": "type va title majburiy"})
	}
	if len(req.Content) == 0 {
		req.Content = json.RawMessage("{}")
	}

	id := uuid.New()
	now := time.Now()
	_, err := h.db.Exec(context.Background(),
		`INSERT INTO board_activities (id, teacher_id, type, title, content, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$6)`,
		id, teacherID, req.Type, req.Title, req.Content, now,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Saqlashda xato"})
	}
	return c.Status(201).JSON(activityDTO{
		ID: id, Type: req.Type, Title: req.Title, Content: req.Content,
		CreatedAt: now, UpdatedAt: now,
	})
}

// PUT /api/activities/:id
func (h *ActivityHandler) Update(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	var req activityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if len(req.Content) == 0 {
		req.Content = json.RawMessage("{}")
	}
	ct, err := h.db.Exec(context.Background(),
		`UPDATE board_activities SET title=$1, content=$2, updated_at=NOW()
		 WHERE id=$3 AND teacher_id=$4`,
		req.Title, req.Content, id, teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Yangilashda xato"})
	}
	if ct.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Topshiriq topilmadi"})
	}
	return c.JSON(fiber.Map{"message": "Saqlandi"})
}

// DELETE /api/activities/:id
func (h *ActivityHandler) Delete(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	_, err = h.db.Exec(context.Background(),
		`DELETE FROM board_activities WHERE id=$1 AND teacher_id=$2`, id, teacherID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "O'chirishda xato"})
	}
	return c.JSON(fiber.Map{"message": "O'chirildi"})
}
