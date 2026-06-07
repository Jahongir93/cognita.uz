package handlers

import (
	"context"
	"math/rand"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gogame.uz/backend/internal/middleware"
	"gogame.uz/backend/internal/models"
)

type ClassHandler struct {
	db *pgxpool.Pool
}

func NewClassHandler(db *pgxpool.Pool) *ClassHandler {
	return &ClassHandler{db: db}
}

// GET /api/classes
func (h *ClassHandler) List(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)

	rows, err := h.db.Query(context.Background(),
		`SELECT c.id, c.teacher_id, c.name, c.grade, c.subject, c.class_code, c.is_active, c.created_at,
		        COUNT(cs.student_id)::int AS student_count
		 FROM classes c
		 LEFT JOIN class_students cs ON cs.class_id = c.id
		 WHERE c.teacher_id = $1
		 GROUP BY c.id
		 ORDER BY c.created_at DESC`,
		teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	classes := make([]models.Class, 0)
	for rows.Next() {
		var cl models.Class
		if err := rows.Scan(&cl.ID, &cl.TeacherID, &cl.Name, &cl.Grade, &cl.Subject,
			&cl.ClassCode, &cl.IsActive, &cl.CreatedAt, &cl.StudentCount); err != nil {
			continue
		}
		classes = append(classes, cl)
	}

	return c.JSON(classes)
}

// POST /api/classes
func (h *ClassHandler) Create(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)

	var req struct {
		Name    string `json:"name"`
		Grade   string `json:"grade"`
		Subject string `json:"subject"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name is required"})
	}

	code := generateClassCode()
	id := uuid.New()

	_, err := h.db.Exec(context.Background(),
		`INSERT INTO classes (id, teacher_id, name, grade, subject, class_code)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, teacherID, req.Name, req.Grade, req.Subject, code,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create class"})
	}

	return c.Status(201).JSON(fiber.Map{"id": id, "class_code": code})
}

// PUT /api/classes/:id
func (h *ClassHandler) Update(c *fiber.Ctx) error {
	classID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}
	teacherID := middleware.GetUserID(c)

	var req struct {
		Name    string `json:"name"`
		Grade   string `json:"grade"`
		Subject string `json:"subject"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	_, err = h.db.Exec(context.Background(),
		`UPDATE classes SET name=$1, grade=$2, subject=$3 WHERE id=$4 AND teacher_id=$5`,
		req.Name, req.Grade, req.Subject, classID, teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Update failed"})
	}

	return c.JSON(fiber.Map{"message": "Updated"})
}

// DELETE /api/classes/:id
func (h *ClassHandler) Delete(c *fiber.Ctx) error {
	classID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}
	teacherID := middleware.GetUserID(c)

	_, err = h.db.Exec(context.Background(),
		`DELETE FROM classes WHERE id=$1 AND teacher_id=$2`,
		classID, teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Delete failed"})
	}

	return c.Status(204).Send(nil)
}

func generateClassCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}
