package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gogame.uz/backend/internal/middleware"
	"gogame.uz/backend/internal/models"
)

type QuizHandler struct {
	db *pgxpool.Pool
}

func NewQuizHandler(db *pgxpool.Pool) *QuizHandler {
	return &QuizHandler{db: db}
}

// GET /api/quizzes  — teacher's own quizzes
func (h *QuizHandler) List(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)

	rows, err := h.db.Query(context.Background(),
		`SELECT id, title, description, subject, grade_level, cover_image,
		        template, is_public, total_questions, play_count, tags, created_at, updated_at
		 FROM quizzes WHERE teacher_id = $1 ORDER BY updated_at DESC`,
		teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	quizzes := make([]models.Quiz, 0)
	for rows.Next() {
		var q models.Quiz
		if err := rows.Scan(&q.ID, &q.Title, &q.Description, &q.Subject, &q.GradeLevel,
			&q.CoverImage, &q.Template, &q.IsPublic, &q.TotalQuestions,
			&q.PlayCount, &q.Tags, &q.CreatedAt, &q.UpdatedAt); err != nil {
			continue
		}
		quizzes = append(quizzes, q)
	}

	return c.JSON(quizzes)
}

// GET /api/quizzes/:id  — full quiz with questions
func (h *QuizHandler) Get(c *fiber.Ctx) error {
	quizID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid quiz ID"})
	}

	teacherID := middleware.GetUserID(c)

	var quiz models.Quiz
	err = h.db.QueryRow(context.Background(),
		`SELECT id, teacher_id, title, description, subject, grade_level, cover_image,
		        template, is_public, total_questions, play_count, tags, created_at, updated_at
		 FROM quizzes WHERE id = $1 AND (teacher_id = $2 OR is_public = true)`,
		quizID, teacherID,
	).Scan(&quiz.ID, &quiz.TeacherID, &quiz.Title, &quiz.Description, &quiz.Subject,
		&quiz.GradeLevel, &quiz.CoverImage, &quiz.Template, &quiz.IsPublic,
		&quiz.TotalQuestions, &quiz.PlayCount, &quiz.Tags, &quiz.CreatedAt, &quiz.UpdatedAt)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Quiz not found"})
	}

	// Load questions
	quiz.Questions, err = h.loadQuestions(quizID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load questions"})
	}

	return c.JSON(quiz)
}

// POST /api/quizzes
func (h *QuizHandler) Create(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)

	var req models.CreateQuizRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Template == "" {
		req.Template = models.TemplateQuiz
	}

	tx, err := h.db.Begin(context.Background())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Transaction failed"})
	}
	defer tx.Rollback(context.Background())

	quizID := uuid.New()
	_, err = tx.Exec(context.Background(),
		`INSERT INTO quizzes (id, teacher_id, title, description, subject, grade_level,
		                      template, is_public, total_questions, tags)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		quizID, teacherID, req.Title, req.Description, req.Subject, req.GradeLevel,
		req.Template, req.IsPublic, len(req.Questions), req.Tags,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create quiz"})
	}

	for _, qReq := range req.Questions {
		qID := uuid.New()
		if qReq.TimeLimit == 0 {
			qReq.TimeLimit = 20
		}
		if qReq.Points == 0 {
			qReq.Points = 100
		}

		_, err = tx.Exec(context.Background(),
			`INSERT INTO questions (id, quiz_id, order_index, type, question_text,
			                        media_url, media_type, time_limit, points, explanation)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			qID, quizID, qReq.OrderIndex, qReq.Type, qReq.QuestionText,
			qReq.MediaURL, qReq.MediaType, qReq.TimeLimit, qReq.Points, qReq.Explanation,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create question"})
		}

		for _, oReq := range qReq.Options {
			_, err = tx.Exec(context.Background(),
				`INSERT INTO answer_options (id, question_id, option_text, media_url, is_correct, order_index)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				uuid.New(), qID, oReq.OptionText, oReq.MediaURL, oReq.IsCorrect, oReq.OrderIndex,
			)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to create option"})
			}
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Commit failed"})
	}

	return c.Status(201).JSON(fiber.Map{
		"id":      quizID,
		"message": "Quiz created successfully",
	})
}

// PUT /api/quizzes/:id
func (h *QuizHandler) Update(c *fiber.Ctx) error {
	quizID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid quiz ID"})
	}

	teacherID := middleware.GetUserID(c)

	var req models.CreateQuizRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.db.Begin(context.Background())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Transaction failed"})
	}
	defer tx.Rollback(context.Background())

	result, err := tx.Exec(context.Background(),
		`UPDATE quizzes SET title=$1, description=$2, subject=$3, grade_level=$4,
		 is_public=$5, tags=$6, total_questions=$7, updated_at=$8
		 WHERE id=$9 AND teacher_id=$10`,
		req.Title, req.Description, req.Subject, req.GradeLevel,
		req.IsPublic, req.Tags, len(req.Questions), time.Now(), quizID, teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Update failed"})
	}
	if result.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Quiz not found or access denied"})
	}

	// Replace all questions
	if _, err = tx.Exec(context.Background(), `DELETE FROM questions WHERE quiz_id = $1`, quizID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to replace questions"})
	}

	for i, qReq := range req.Questions {
		qReq.OrderIndex = i
		if qReq.TimeLimit == 0 {
			qReq.TimeLimit = 20
		}
		if qReq.Points == 0 {
			qReq.Points = 100
		}
		qID := uuid.New()
		_, err = tx.Exec(context.Background(),
			`INSERT INTO questions (id, quiz_id, order_index, type, question_text,
			                        media_url, media_type, time_limit, points, explanation)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			qID, quizID, qReq.OrderIndex, qReq.Type, qReq.QuestionText,
			qReq.MediaURL, qReq.MediaType, qReq.TimeLimit, qReq.Points, qReq.Explanation,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save question"})
		}
		for j, oReq := range qReq.Options {
			oReq.OrderIndex = j
			_, err = tx.Exec(context.Background(),
				`INSERT INTO answer_options (id, question_id, option_text, media_url, is_correct, order_index)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				uuid.New(), qID, oReq.OptionText, oReq.MediaURL, oReq.IsCorrect, oReq.OrderIndex,
			)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to save option"})
			}
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Commit failed"})
	}

	return c.JSON(fiber.Map{"message": "Quiz updated"})
}

// DELETE /api/quizzes/:id
func (h *QuizHandler) Delete(c *fiber.Ctx) error {
	quizID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid quiz ID"})
	}

	teacherID := middleware.GetUserID(c)

	_, err = h.db.Exec(context.Background(),
		`DELETE FROM quizzes WHERE id=$1 AND teacher_id=$2`,
		quizID, teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Delete failed"})
	}

	return c.JSON(fiber.Map{"message": "Quiz deleted"})
}

func (h *QuizHandler) loadQuestions(quizID uuid.UUID) ([]models.Question, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, quiz_id, order_index, type, question_text,
		        media_url, media_type, time_limit, points, explanation
		 FROM questions WHERE quiz_id = $1 ORDER BY order_index`,
		quizID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := make([]models.Question, 0)
	for rows.Next() {
		var q models.Question
		if err := rows.Scan(&q.ID, &q.QuizID, &q.OrderIndex, &q.Type, &q.QuestionText,
			&q.MediaURL, &q.MediaType, &q.TimeLimit, &q.Points, &q.Explanation); err != nil {
			continue
		}
		questions = append(questions, q)
	}

	// Load options for all questions
	for i := range questions {
		questions[i].Options, err = h.loadOptions(questions[i].ID)
		if err != nil {
			return nil, fmt.Errorf("loading options: %w", err)
		}
	}

	return questions, nil
}

func (h *QuizHandler) loadOptions(questionID uuid.UUID) ([]models.AnswerOption, error) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, question_id, option_text, media_url, is_correct, order_index
		 FROM answer_options WHERE question_id = $1 ORDER BY order_index`,
		questionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]models.AnswerOption, 0)
	for rows.Next() {
		var o models.AnswerOption
		if err := rows.Scan(&o.ID, &o.QuestionID, &o.OptionText,
			&o.MediaURL, &o.IsCorrect, &o.OrderIndex); err != nil {
			continue
		}
		options = append(options, o)
	}
	return options, nil
}

// GET /api/quizzes/discover?q=&subject=&grade=  — browse all public quizzes (any authenticated user)
func (h *QuizHandler) Discover(c *fiber.Ctx) error {
	search := c.Query("q")
	subject := c.Query("subject")
	grade := c.Query("grade")

	query := `SELECT id, teacher_id, title, description, subject, grade_level,
	                 cover_image, template, is_public, total_questions, play_count,
	                 tags, created_at, updated_at
	          FROM quizzes WHERE is_public = true`
	args := []any{}
	idx := 1

	if subject != "" {
		query += fmt.Sprintf(" AND subject = $%d", idx)
		args = append(args, subject)
		idx++
	}
	if grade != "" {
		query += fmt.Sprintf(" AND grade_level = $%d", idx)
		args = append(args, grade)
		idx++
	}
	if search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}
	_ = idx
	query += " ORDER BY play_count DESC, created_at DESC LIMIT 100"

	rows, err := h.db.Query(context.Background(), query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	quizzes := make([]models.Quiz, 0)
	for rows.Next() {
		var q models.Quiz
		if err := rows.Scan(&q.ID, &q.TeacherID, &q.Title, &q.Description, &q.Subject,
			&q.GradeLevel, &q.CoverImage, &q.Template, &q.IsPublic, &q.TotalQuestions,
			&q.PlayCount, &q.Tags, &q.CreatedAt, &q.UpdatedAt); err != nil {
			continue
		}
		quizzes = append(quizzes, q)
	}
	return c.JSON(quizzes)
}
