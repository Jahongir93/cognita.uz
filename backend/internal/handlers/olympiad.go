package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gogame.uz/backend/internal/middleware"
	"gogame.uz/backend/internal/models"
)

// ─── OlympiadHandler ─────────────────────────────────────────────────────────

type OlympiadHandler struct {
	db *pgxpool.Pool
}

func NewOlympiadHandler(db *pgxpool.Pool) *OlympiadHandler {
	return &OlympiadHandler{db: db}
}

// GET /api/olympiads
func (h *OlympiadHandler) List(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)

	rows, err := h.db.Query(context.Background(),
		`SELECT o.id, o.quiz_id, o.teacher_id, o.title, o.description, o.code,
		        o.start_time, o.end_time, o.max_participants, o.status, o.created_at,
		        COALESCE(q.title, '') AS quiz_title,
		        COUNT(os.id) AS participant_count
		 FROM olympiads o
		 LEFT JOIN quizzes q ON q.id = o.quiz_id
		 LEFT JOIN olympiad_submissions os ON os.olympiad_id = o.id
		 WHERE o.teacher_id = $1
		 GROUP BY o.id, q.title
		 ORDER BY o.created_at DESC`,
		teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	olympiads := make([]models.Olympiad, 0)
	for rows.Next() {
		var ol models.Olympiad
		if err := rows.Scan(
			&ol.ID, &ol.QuizID, &ol.TeacherID, &ol.Title, &ol.Description, &ol.Code,
			&ol.StartTime, &ol.EndTime, &ol.MaxParticipants, &ol.Status, &ol.CreatedAt,
			&ol.QuizTitle, &ol.ParticipantCount,
		); err != nil {
			continue
		}
		olympiads = append(olympiads, ol)
	}
	return c.JSON(olympiads)
}

// POST /api/olympiads
func (h *OlympiadHandler) Create(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)

	var req models.CreateOlympiadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Title is required"})
	}
	if req.QuizID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "quiz_id is required"})
	}

	quizID, err := uuid.Parse(req.QuizID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid quiz_id"})
	}

	if req.EndTime.Before(req.StartTime) {
		return c.Status(400).JSON(fiber.Map{"error": "end_time must be after start_time"})
	}

	code, err := generateCode(8)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate code"})
	}

	var olympiadID uuid.UUID
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO olympiads (quiz_id, teacher_id, title, description, code,
		                        start_time, end_time, max_participants, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'upcoming')
		 RETURNING id`,
		quizID, teacherID, req.Title, req.Description, code,
		req.StartTime, req.EndTime, req.MaxParticipants,
	).Scan(&olympiadID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create olympiad"})
	}

	return c.Status(201).JSON(fiber.Map{
		"id":   olympiadID,
		"code": code,
	})
}

// GET /api/olympiads/:id
func (h *OlympiadHandler) Get(c *fiber.Ctx) error {
	olympiadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid olympiad ID"})
	}
	teacherID := middleware.GetUserID(c)

	var ol models.Olympiad
	err = h.db.QueryRow(context.Background(),
		`SELECT o.id, o.quiz_id, o.teacher_id, o.title, o.description, o.code,
		        o.start_time, o.end_time, o.max_participants, o.status, o.created_at,
		        COALESCE(q.title, '') AS quiz_title,
		        COUNT(os.id) AS participant_count
		 FROM olympiads o
		 LEFT JOIN quizzes q ON q.id = o.quiz_id
		 LEFT JOIN olympiad_submissions os ON os.olympiad_id = o.id
		 WHERE o.id = $1 AND o.teacher_id = $2
		 GROUP BY o.id, q.title`,
		olympiadID, teacherID,
	).Scan(
		&ol.ID, &ol.QuizID, &ol.TeacherID, &ol.Title, &ol.Description, &ol.Code,
		&ol.StartTime, &ol.EndTime, &ol.MaxParticipants, &ol.Status, &ol.CreatedAt,
		&ol.QuizTitle, &ol.ParticipantCount,
	)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Olympiad not found"})
	}
	return c.JSON(ol)
}

// PUT /api/olympiads/:id
func (h *OlympiadHandler) Update(c *fiber.Ctx) error {
	olympiadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid olympiad ID"})
	}
	teacherID := middleware.GetUserID(c)

	var req models.CreateOlympiadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.db.Exec(context.Background(),
		`UPDATE olympiads SET title=$1, description=$2, start_time=$3,
		                      end_time=$4, max_participants=$5
		 WHERE id=$6 AND teacher_id=$7`,
		req.Title, req.Description, req.StartTime, req.EndTime,
		req.MaxParticipants, olympiadID, teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Update failed"})
	}
	if result.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Olympiad not found or access denied"})
	}
	return c.JSON(fiber.Map{"message": "Olympiad updated"})
}

// PUT /api/olympiads/:id/status
func (h *OlympiadHandler) SetStatus(c *fiber.Ctx) error {
	olympiadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid olympiad ID"})
	}
	teacherID := middleware.GetUserID(c)

	var body struct {
		Status models.OlympiadStatus `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	switch body.Status {
	case models.OlympiadUpcoming, models.OlympiadActive, models.OlympiadCompleted:
		// valid
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Invalid status value"})
	}

	result, err := h.db.Exec(context.Background(),
		`UPDATE olympiads SET status=$1 WHERE id=$2 AND teacher_id=$3`,
		body.Status, olympiadID, teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Update failed"})
	}
	if result.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Olympiad not found or access denied"})
	}
	return c.JSON(fiber.Map{"status": body.Status})
}

// DELETE /api/olympiads/:id
func (h *OlympiadHandler) Delete(c *fiber.Ctx) error {
	olympiadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid olympiad ID"})
	}
	teacherID := middleware.GetUserID(c)

	_, err = h.db.Exec(context.Background(),
		`DELETE FROM olympiads WHERE id=$1 AND teacher_id=$2`,
		olympiadID, teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Delete failed"})
	}
	return c.SendStatus(204)
}

// GET /api/olympiads/join/:code  — PUBLIC
func (h *OlympiadHandler) JoinByCode(c *fiber.Ctx) error {
	code := c.Params("code")

	var ol models.Olympiad
	var participantCount int
	err := h.db.QueryRow(context.Background(),
		`SELECT o.id, o.quiz_id, o.teacher_id, o.title, o.description, o.code,
		        o.start_time, o.end_time, o.max_participants, o.status, o.created_at,
		        COALESCE(q.title, '') AS quiz_title,
		        COUNT(os.id) AS participant_count
		 FROM olympiads o
		 LEFT JOIN quizzes q ON q.id = o.quiz_id
		 LEFT JOIN olympiad_submissions os ON os.olympiad_id = o.id
		 WHERE o.code = $1
		 GROUP BY o.id, q.title`,
		code,
	).Scan(
		&ol.ID, &ol.QuizID, &ol.TeacherID, &ol.Title, &ol.Description, &ol.Code,
		&ol.StartTime, &ol.EndTime, &ol.MaxParticipants, &ol.Status, &ol.CreatedAt,
		&ol.QuizTitle, &participantCount,
	)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Olympiad not found"})
	}

	now := time.Now()
	isActive := ol.Status == models.OlympiadActive &&
		now.After(ol.StartTime) && now.Before(ol.EndTime)

	return c.JSON(fiber.Map{
		"olympiad":          ol,
		"participant_count": participantCount,
		"is_active":         isActive,
	})
}

// GET /api/olympiads/take/:code  — PUBLIC
func (h *OlympiadHandler) TakeOlympiad(c *fiber.Ctx) error {
	code := c.Params("code")

	var ol models.Olympiad
	err := h.db.QueryRow(context.Background(),
		`SELECT id, quiz_id, title, code, start_time, end_time, status
		 FROM olympiads WHERE code = $1`,
		code,
	).Scan(&ol.ID, &ol.QuizID, &ol.Title, &ol.Code, &ol.StartTime, &ol.EndTime, &ol.Status)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Olympiad not found"})
	}

	if ol.Status != models.OlympiadActive {
		return c.Status(403).JSON(fiber.Map{"error": "Olympiad is not active"})
	}

	now := time.Now()
	if now.Before(ol.StartTime) {
		return c.Status(403).JSON(fiber.Map{"error": "Olympiad has not started yet"})
	}
	if now.After(ol.EndTime) {
		return c.Status(403).JSON(fiber.Map{"error": "Olympiad has already ended"})
	}

	questions, err := loadQuestionsFromDB(h.db, ol.QuizID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load questions"})
	}

	// Strip IsCorrect from options
	type studentQuestion struct {
		ID           uuid.UUID                    `json:"id"`
		OrderIndex   int                          `json:"order_index"`
		Type         models.QuestionType          `json:"type"`
		QuestionText string                       `json:"question_text"`
		MediaURL     *string                      `json:"media_url"`
		MediaType    *string                      `json:"media_type"`
		TimeLimit    int                          `json:"time_limit"`
		Points       int                          `json:"points"`
		Options      []models.StudentAnswerOption `json:"options"`
	}

	result := make([]studentQuestion, 0, len(questions))
	for _, q := range questions {
		sq := studentQuestion{
			ID:           q.ID,
			OrderIndex:   q.OrderIndex,
			Type:         q.Type,
			QuestionText: q.QuestionText,
			MediaURL:     q.MediaURL,
			MediaType:    q.MediaType,
			TimeLimit:    q.TimeLimit,
			Points:       q.Points,
		}
		for i := range q.Options {
			sq.Options = append(sq.Options, q.Options[i].ToStudent())
		}
		result = append(result, sq)
	}

	return c.JSON(fiber.Map{
		"olympiad":  ol,
		"questions": result,
	})
}

// POST /api/olympiads/submit/:code  — PUBLIC
func (h *OlympiadHandler) Submit(c *fiber.Ctx) error {
	code := c.Params("code")

	var ol models.Olympiad
	err := h.db.QueryRow(context.Background(),
		`SELECT id, quiz_id, status, start_time, end_time, max_participants FROM olympiads WHERE code = $1`,
		code,
	).Scan(&ol.ID, &ol.QuizID, &ol.Status, &ol.StartTime, &ol.EndTime, &ol.MaxParticipants)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Olympiad not found"})
	}

	if ol.Status != models.OlympiadActive {
		return c.Status(403).JSON(fiber.Map{"error": "Olympiad is not active"})
	}

	now := time.Now()
	if now.Before(ol.StartTime) {
		return c.Status(403).JSON(fiber.Map{"error": "Olympiad has not started yet"})
	}
	if now.After(ol.EndTime) {
		return c.Status(403).JSON(fiber.Map{"error": "Olympiad has already ended"})
	}

	// Check max_participants cap
	if ol.MaxParticipants != nil {
		var count int
		_ = h.db.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM olympiad_submissions WHERE olympiad_id = $1`, ol.ID,
		).Scan(&count)
		if count >= *ol.MaxParticipants {
			return c.Status(403).JSON(fiber.Map{"error": "Olympiad is full"})
		}
	}

	var body struct {
		StudentName string                   `json:"student_name"`
		Answers     []models.SubmittedAnswer `json:"answers"`
		TimeTaken   int                      `json:"time_taken"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.StudentName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "student_name is required"})
	}

	score, maxScore, err := scoreAnswers(h.db, ol.QuizID, body.Answers)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to score answers"})
	}

	answersJSON, err := json.Marshal(body.Answers)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to encode answers"})
	}

	var submissionID uuid.UUID
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO olympiad_submissions (olympiad_id, student_name, answers, score, max_score, time_taken)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		ol.ID, body.StudentName, answersJSON, score, maxScore, body.TimeTaken,
	).Scan(&submissionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save submission"})
	}

	percent := 0
	if maxScore > 0 {
		percent = score * 100 / maxScore
	}

	return c.Status(201).JSON(fiber.Map{
		"id":        submissionID,
		"score":     score,
		"max_score": maxScore,
		"percent":   percent,
	})
}

// GET /api/olympiads/:id/leaderboard  — PUBLIC
func (h *OlympiadHandler) Leaderboard(c *fiber.Ctx) error {
	olympiadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid olympiad ID"})
	}

	// Verify olympiad exists
	var exists bool
	_ = h.db.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM olympiads WHERE id = $1)`, olympiadID,
	).Scan(&exists)
	if !exists {
		return c.Status(404).JSON(fiber.Map{"error": "Olympiad not found"})
	}

	rows, err := h.db.Query(context.Background(),
		`SELECT id, olympiad_id, student_name, answers, score, max_score, time_taken, submitted_at
		 FROM olympiad_submissions
		 WHERE olympiad_id = $1
		 ORDER BY score DESC, time_taken ASC`,
		olympiadID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	submissions := make([]models.OlympiadSubmission, 0)
	rank := 1
	for rows.Next() {
		var sub models.OlympiadSubmission
		var answersJSON []byte
		if err := rows.Scan(
			&sub.ID, &sub.OlympiadID, &sub.StudentName, &answersJSON,
			&sub.Score, &sub.MaxScore, &sub.TimeTaken, &sub.SubmittedAt,
		); err != nil {
			continue
		}
		if err := json.Unmarshal(answersJSON, &sub.Answers); err != nil {
			sub.Answers = []models.SubmittedAnswer{}
		}
		sub.Rank = rank
		rank++
		submissions = append(submissions, sub)
	}

	return c.JSON(submissions)
}
