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

// ─── Talaba tomoni ───────────────────────────────────────────────────────────

// POST /api/classes/join  — talaba kod orqali sinfga qo'shiladi
func (h *ClassHandler) Join(c *fiber.Ctx) error {
	studentID := middleware.GetUserID(c)
	var req struct {
		ClassCode string `json:"class_code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	code := ""
	for _, r := range req.ClassCode {
		if r != ' ' {
			if r >= 'a' && r <= 'z' {
				r -= 32
			}
			code += string(r)
		}
	}
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Sinf kodini kiriting"})
	}

	var classID uuid.UUID
	var name string
	err := h.db.QueryRow(context.Background(),
		`SELECT id, name FROM classes WHERE class_code=$1 AND is_active=true`, code,
	).Scan(&classID, &name)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Bunday kodli sinf topilmadi"})
	}

	_, err = h.db.Exec(context.Background(),
		`INSERT INTO class_students (class_id, student_id) VALUES ($1, $2)
		 ON CONFLICT (class_id, student_id) DO NOTHING`,
		classID, studentID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Qo'shilishda xato"})
	}
	return c.JSON(fiber.Map{"class_id": classID, "name": name, "message": "Sinfga qo'shildingiz"})
}

// GET /api/classes/my  — talabaning qo'shilgan sinflari
func (h *ClassHandler) MyClasses(c *fiber.Ctx) error {
	studentID := middleware.GetUserID(c)
	rows, err := h.db.Query(context.Background(),
		`SELECT c.id, c.name, c.grade, c.subject, c.class_code, c.created_at,
		        COALESCE(u.full_name, '') AS teacher_name,
		        COUNT(cs2.student_id)::int AS student_count
		 FROM class_students cs
		 JOIN classes c ON c.id = cs.class_id
		 LEFT JOIN users u ON u.id = c.teacher_id
		 LEFT JOIN class_students cs2 ON cs2.class_id = c.id
		 WHERE cs.student_id = $1
		 GROUP BY c.id, u.full_name
		 ORDER BY c.created_at DESC`,
		studentID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	type myClass struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		Grade        *string   `json:"grade"`
		Subject      *string   `json:"subject"`
		ClassCode    string    `json:"class_code"`
		CreatedAt    any       `json:"created_at"`
		TeacherName  string    `json:"teacher_name"`
		StudentCount int       `json:"student_count"`
	}
	list := make([]myClass, 0)
	for rows.Next() {
		var m myClass
		if err := rows.Scan(&m.ID, &m.Name, &m.Grade, &m.Subject, &m.ClassCode,
			&m.CreatedAt, &m.TeacherName, &m.StudentCount); err != nil {
			continue
		}
		list = append(list, m)
	}
	return c.JSON(list)
}

// POST /api/classes/:id/leave  — talaba sinfdan chiqadi
func (h *ClassHandler) Leave(c *fiber.Ctx) error {
	studentID := middleware.GetUserID(c)
	classID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}
	_, err = h.db.Exec(context.Background(),
		`DELETE FROM class_students WHERE class_id=$1 AND student_id=$2`, classID, studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Chiqishda xato"})
	}
	return c.Status(204).Send(nil)
}

// GET /api/classes/:id/students  — o'qituvchi: sinf o'quvchilari
func (h *ClassHandler) Students(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)
	classID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}
	// Egalikni tekshirish
	var owner uuid.UUID
	if err := h.db.QueryRow(context.Background(),
		`SELECT teacher_id FROM classes WHERE id=$1`, classID).Scan(&owner); err != nil || owner != teacherID {
		return c.Status(403).JSON(fiber.Map{"error": "Ruxsat yo'q"})
	}

	rows, err := h.db.Query(context.Background(),
		`SELECT u.id, u.full_name, u.username, u.email, cs.joined_at
		 FROM class_students cs JOIN users u ON u.id = cs.student_id
		 WHERE cs.class_id = $1 ORDER BY cs.joined_at DESC`, classID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	type student struct {
		ID       uuid.UUID `json:"id"`
		FullName string    `json:"full_name"`
		Username string    `json:"username"`
		Email    string    `json:"email"`
		JoinedAt any       `json:"joined_at"`
	}
	list := make([]student, 0)
	for rows.Next() {
		var s student
		if err := rows.Scan(&s.ID, &s.FullName, &s.Username, &s.Email, &s.JoinedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return c.JSON(list)
}

func generateClassCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}
