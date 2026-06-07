package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gogame.uz/backend/internal/middleware"
	"gogame.uz/backend/internal/models"
)

// codeChars excludes ambiguous characters: 0, O, 1, I
const codeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// generateCode returns a cryptographically random uppercase alphanumeric code
// of length n, excluding ambiguous characters (0, O, 1, I).
func generateCode(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	result := make([]byte, n)
	for i, b := range buf {
		result[i] = codeChars[int(b)%len(codeChars)]
	}
	return string(result), nil
}

// loadQuestionsFromDB loads all questions (with options) for a quiz.
// Exported as a package-level function so it can be reused by olympiad.go.
func loadQuestionsFromDB(db *pgxpool.Pool, quizID uuid.UUID) ([]models.Question, error) {
	rows, err := db.Query(context.Background(),
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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range questions {
		questions[i].Options, err = loadOptionsFromDB(db, questions[i].ID)
		if err != nil {
			return nil, fmt.Errorf("loading options: %w", err)
		}
	}
	return questions, nil
}

// loadOptionsFromDB loads all answer options for a question.
func loadOptionsFromDB(db *pgxpool.Pool, questionID uuid.UUID) ([]models.AnswerOption, error) {
	rows, err := db.Query(context.Background(),
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
	return options, rows.Err()
}

// scoreAnswers computes score and maxScore for a set of submitted answers
// against all answer_options for the given quiz.
func scoreAnswers(db *pgxpool.Pool, quizID uuid.UUID, submitted []models.SubmittedAnswer) (score, maxScore int, err error) {
	// Load all questions to determine per-question max points
	questions, err := loadQuestionsFromDB(db, quizID)
	if err != nil {
		return 0, 0, err
	}

	// Build option map: optionID -> {isCorrect, points}
	type optInfo struct {
		isCorrect bool
		points    int
	}
	optMap := make(map[string]optInfo)
	for _, q := range questions {
		maxScore += q.Points
		for _, o := range q.Options {
			optMap[o.ID.String()] = optInfo{isCorrect: o.IsCorrect, points: q.Points}
		}
	}

	// Score submitted answers
	for _, sa := range submitted {
		if sa.OptionID == "" {
			continue
		}
		if info, ok := optMap[sa.OptionID]; ok && info.isCorrect {
			score += info.points
		}
	}
	return score, maxScore, nil
}

// ─── ExamHandler ─────────────────────────────────────────────────────────────

type ExamHandler struct {
	db *pgxpool.Pool
}

func NewExamHandler(db *pgxpool.Pool) *ExamHandler {
	return &ExamHandler{db: db}
}

// GET /api/exams
func (h *ExamHandler) List(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)

	rows, err := h.db.Query(context.Background(),
		`SELECT e.id, e.quiz_id, e.teacher_id, e.title, e.code, e.time_limit,
		        e.start_date, e.end_date, e.shuffle, e.max_attempts, e.status, e.created_at,
		        COALESCE(q.title, '') AS quiz_title,
		        COUNT(es.id) AS submission_count
		 FROM exams e
		 LEFT JOIN quizzes q ON q.id = e.quiz_id
		 LEFT JOIN exam_submissions es ON es.exam_id = e.id
		 WHERE e.teacher_id = $1
		 GROUP BY e.id, q.title
		 ORDER BY e.created_at DESC`,
		teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	exams := make([]models.Exam, 0)
	for rows.Next() {
		var ex models.Exam
		if err := rows.Scan(
			&ex.ID, &ex.QuizID, &ex.TeacherID, &ex.Title, &ex.Code, &ex.TimeLimit,
			&ex.StartDate, &ex.EndDate, &ex.Shuffle, &ex.MaxAttempts, &ex.Status, &ex.CreatedAt,
			&ex.QuizTitle, &ex.SubmissionCount,
		); err != nil {
			continue
		}
		exams = append(exams, ex)
	}
	return c.JSON(exams)
}

// POST /api/exams
func (h *ExamHandler) Create(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)

	var req models.CreateExamRequest
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

	if req.TimeLimit <= 0 {
		req.TimeLimit = 30
	}
	if req.MaxAttempts <= 0 {
		req.MaxAttempts = 1
	}

	code, err := generateCode(6)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate code"})
	}

	var examID uuid.UUID
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO exams (quiz_id, teacher_id, title, code, time_limit, start_date, end_date,
		                    shuffle, max_attempts, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'draft')
		 RETURNING id`,
		quizID, teacherID, req.Title, code, req.TimeLimit,
		req.StartDate, req.EndDate, req.Shuffle, req.MaxAttempts,
	).Scan(&examID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create exam"})
	}

	return c.Status(201).JSON(fiber.Map{
		"id":   examID,
		"code": code,
	})
}

// GET /api/exams/:id
func (h *ExamHandler) Get(c *fiber.Ctx) error {
	examID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid exam ID"})
	}
	teacherID := middleware.GetUserID(c)

	var ex models.Exam
	err = h.db.QueryRow(context.Background(),
		`SELECT e.id, e.quiz_id, e.teacher_id, e.title, e.code, e.time_limit,
		        e.start_date, e.end_date, e.shuffle, e.max_attempts, e.status, e.created_at,
		        COALESCE(q.title, '') AS quiz_title,
		        COUNT(es.id) AS submission_count
		 FROM exams e
		 LEFT JOIN quizzes q ON q.id = e.quiz_id
		 LEFT JOIN exam_submissions es ON es.exam_id = e.id
		 WHERE e.id = $1 AND e.teacher_id = $2
		 GROUP BY e.id, q.title`,
		examID, teacherID,
	).Scan(
		&ex.ID, &ex.QuizID, &ex.TeacherID, &ex.Title, &ex.Code, &ex.TimeLimit,
		&ex.StartDate, &ex.EndDate, &ex.Shuffle, &ex.MaxAttempts, &ex.Status, &ex.CreatedAt,
		&ex.QuizTitle, &ex.SubmissionCount,
	)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Exam not found"})
	}
	return c.JSON(ex)
}

// PUT /api/exams/:id
func (h *ExamHandler) Update(c *fiber.Ctx) error {
	examID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid exam ID"})
	}
	teacherID := middleware.GetUserID(c)

	var req models.CreateExamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.db.Exec(context.Background(),
		`UPDATE exams SET title=$1, time_limit=$2, start_date=$3, end_date=$4,
		                  shuffle=$5, max_attempts=$6
		 WHERE id=$7 AND teacher_id=$8`,
		req.Title, req.TimeLimit, req.StartDate, req.EndDate,
		req.Shuffle, req.MaxAttempts, examID, teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Update failed"})
	}
	if result.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Exam not found or access denied"})
	}
	return c.JSON(fiber.Map{"message": "Exam updated"})
}

// PUT /api/exams/:id/status
func (h *ExamHandler) SetStatus(c *fiber.Ctx) error {
	examID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid exam ID"})
	}
	teacherID := middleware.GetUserID(c)

	var body struct {
		Status models.ExamStatus `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	switch body.Status {
	case models.ExamDraft, models.ExamActive, models.ExamClosed:
		// valid
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Invalid status value"})
	}

	result, err := h.db.Exec(context.Background(),
		`UPDATE exams SET status=$1 WHERE id=$2 AND teacher_id=$3`,
		body.Status, examID, teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Update failed"})
	}
	if result.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Exam not found or access denied"})
	}
	return c.JSON(fiber.Map{"status": body.Status})
}

// DELETE /api/exams/:id
func (h *ExamHandler) Delete(c *fiber.Ctx) error {
	examID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid exam ID"})
	}
	teacherID := middleware.GetUserID(c)

	_, err = h.db.Exec(context.Background(),
		`DELETE FROM exams WHERE id=$1 AND teacher_id=$2`,
		examID, teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Delete failed"})
	}
	return c.SendStatus(204)
}

// GET /api/exams/join/:code  — PUBLIC
func (h *ExamHandler) JoinByCode(c *fiber.Ctx) error {
	code := c.Params("code")

	var ex models.Exam
	err := h.db.QueryRow(context.Background(),
		`SELECT e.id, e.quiz_id, e.teacher_id, e.title, e.code, e.time_limit,
		        e.start_date, e.end_date, e.shuffle, e.max_attempts, e.status, e.created_at,
		        COALESCE(q.title, '') AS quiz_title
		 FROM exams e
		 LEFT JOIN quizzes q ON q.id = e.quiz_id
		 WHERE e.code = $1`,
		code,
	).Scan(
		&ex.ID, &ex.QuizID, &ex.TeacherID, &ex.Title, &ex.Code, &ex.TimeLimit,
		&ex.StartDate, &ex.EndDate, &ex.Shuffle, &ex.MaxAttempts, &ex.Status, &ex.CreatedAt,
		&ex.QuizTitle,
	)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Exam not found"})
	}

	if ex.Status != models.ExamActive {
		return c.Status(403).JSON(fiber.Map{"error": "Exam is not active"})
	}

	now := time.Now()
	if ex.StartDate != nil && now.Before(*ex.StartDate) {
		return c.Status(403).JSON(fiber.Map{"error": "Exam has not started yet"})
	}
	if ex.EndDate != nil && now.After(*ex.EndDate) {
		return c.Status(403).JSON(fiber.Map{"error": "Exam has already ended"})
	}

	return c.JSON(ex)
}

// GET /api/exams/take/:code  — PUBLIC
func (h *ExamHandler) TakeExam(c *fiber.Ctx) error {
	code := c.Params("code")

	var ex models.Exam
	err := h.db.QueryRow(context.Background(),
		`SELECT id, quiz_id, title, code, time_limit, start_date, end_date, shuffle, status
		 FROM exams WHERE code = $1`,
		code,
	).Scan(
		&ex.ID, &ex.QuizID, &ex.Title, &ex.Code, &ex.TimeLimit,
		&ex.StartDate, &ex.EndDate, &ex.Shuffle, &ex.Status,
	)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Exam not found"})
	}

	if ex.Status != models.ExamActive {
		return c.Status(403).JSON(fiber.Map{"error": "Exam is not active"})
	}

	now := time.Now()
	if ex.StartDate != nil && now.Before(*ex.StartDate) {
		return c.Status(403).JSON(fiber.Map{"error": "Exam has not started yet"})
	}
	if ex.EndDate != nil && now.After(*ex.EndDate) {
		return c.Status(403).JSON(fiber.Map{"error": "Exam has already ended"})
	}

	questions, err := loadQuestionsFromDB(h.db, ex.QuizID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load questions"})
	}

	// Return questions with options stripped of IsCorrect
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
		"exam":      ex,
		"questions": result,
	})
}

// POST /api/exams/submit/:code  — PUBLIC
func (h *ExamHandler) Submit(c *fiber.Ctx) error {
	code := c.Params("code")

	var ex models.Exam
	err := h.db.QueryRow(context.Background(),
		`SELECT id, quiz_id, status, start_date, end_date FROM exams WHERE code = $1`,
		code,
	).Scan(&ex.ID, &ex.QuizID, &ex.Status, &ex.StartDate, &ex.EndDate)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Exam not found"})
	}

	if ex.Status != models.ExamActive {
		return c.Status(403).JSON(fiber.Map{"error": "Exam is not active"})
	}

	now := time.Now()
	if ex.StartDate != nil && now.Before(*ex.StartDate) {
		return c.Status(403).JSON(fiber.Map{"error": "Exam has not started yet"})
	}
	if ex.EndDate != nil && now.After(*ex.EndDate) {
		return c.Status(403).JSON(fiber.Map{"error": "Exam has already ended"})
	}

	var body struct {
		StudentName string                    `json:"student_name"`
		Answers     []models.SubmittedAnswer  `json:"answers"`
		TimeTaken   int                       `json:"time_taken"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.StudentName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "student_name is required"})
	}

	score, maxScore, err := scoreAnswers(h.db, ex.QuizID, body.Answers)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to score answers"})
	}

	answersJSON, err := json.Marshal(body.Answers)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to encode answers"})
	}

	var submissionID uuid.UUID
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO exam_submissions (exam_id, student_name, answers, score, max_score, time_taken)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		ex.ID, body.StudentName, answersJSON, score, maxScore, body.TimeTaken,
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

// GET /api/exams/:id/results  — teacher only
func (h *ExamHandler) Results(c *fiber.Ctx) error {
	examID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid exam ID"})
	}
	teacherID := middleware.GetUserID(c)

	// Verify ownership
	var ownerID uuid.UUID
	err = h.db.QueryRow(context.Background(),
		`SELECT teacher_id FROM exams WHERE id = $1`, examID,
	).Scan(&ownerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Exam not found"})
	}
	if ownerID != teacherID {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	rows, err := h.db.Query(context.Background(),
		`SELECT id, exam_id, student_name, answers, score, max_score, time_taken, submitted_at
		 FROM exam_submissions
		 WHERE exam_id = $1
		 ORDER BY score DESC, time_taken ASC`,
		examID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	submissions := make([]models.ExamSubmission, 0)
	for rows.Next() {
		var sub models.ExamSubmission
		var answersJSON []byte
		if err := rows.Scan(
			&sub.ID, &sub.ExamID, &sub.StudentName, &answersJSON,
			&sub.Score, &sub.MaxScore, &sub.TimeTaken, &sub.SubmittedAt,
		); err != nil {
			continue
		}
		if err := json.Unmarshal(answersJSON, &sub.Answers); err != nil {
			sub.Answers = []models.SubmittedAnswer{}
		}
		submissions = append(submissions, sub)
	}

	return c.JSON(submissions)
}
