package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	fws "github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gogame.uz/backend/internal/game"
	"gogame.uz/backend/internal/middleware"
	"gogame.uz/backend/internal/models"
	ws "gogame.uz/backend/internal/websocket"
)

type RoomHandler struct {
	db  *pgxpool.Pool
	hub *ws.Hub
}

func NewRoomHandler(db *pgxpool.Pool, hub *ws.Hub) *RoomHandler {
	h := &RoomHandler{db: db, hub: hub}
	// Wire up student join handler so the Hub can call it with DB access
	hub.JoinHandler = func(c *ws.Client, payload ws.JoinRoomPayload) error {
		return h.handleStudentJoin(c, payload.PIN, payload.Nickname, payload.Avatar)
	}
	return h
}

// POST /api/rooms  — teacher creates a new game room
func (h *RoomHandler) Create(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)

	var req models.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.GameMode == "" {
		req.GameMode = models.ModeClassic
	}

	// Verify quiz belongs to teacher
	var quizTeacherID uuid.UUID
	err := h.db.QueryRow(context.Background(),
		`SELECT teacher_id FROM quizzes WHERE id = $1`,
		req.QuizID,
	).Scan(&quizTeacherID)
	if err != nil || quizTeacherID != teacherID {
		return c.Status(403).JSON(fiber.Map{"error": "Quiz not found or access denied"})
	}

	// Generate unique PIN
	pin, err := h.generatePIN()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate PIN"})
	}

	settingsJSON, _ := json.Marshal(req.Settings)
	roomID := uuid.New()

	_, err = h.db.Exec(context.Background(),
		`INSERT INTO rooms (id, quiz_id, host_id, pin, game_mode, settings)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		roomID, req.QuizID, teacherID, pin, req.GameMode, settingsJSON,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create room"})
	}

	return c.Status(201).JSON(fiber.Map{
		"room_id": roomID,
		"pin":     pin,
	})
}

// GET /api/rooms/history  — teacher's past game sessions
func (h *RoomHandler) History(c *fiber.Ctx) error {
	teacherID := middleware.GetUserID(c)

	rows, err := h.db.Query(context.Background(),
		`SELECT r.id, r.pin, r.quiz_id, q.title, r.game_mode::text, r.status::text,
		        r.created_at, r.ended_at,
		        COUNT(DISTINCT rp.id)::int AS player_count
		 FROM rooms r
		 JOIN quizzes q ON q.id = r.quiz_id
		 LEFT JOIN room_participants rp ON rp.room_id = r.id
		 WHERE r.host_id = $1
		 GROUP BY r.id, q.title
		 ORDER BY r.created_at DESC
		 LIMIT 50`,
		teacherID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	type RoomHistory struct {
		ID          uuid.UUID  `json:"id"`
		PIN         string     `json:"pin"`
		QuizID      uuid.UUID  `json:"quiz_id"`
		QuizTitle   string     `json:"quiz_title"`
		GameMode    string     `json:"game_mode"`
		Status      string     `json:"status"`
		CreatedAt   time.Time  `json:"created_at"`
		EndedAt     *time.Time `json:"ended_at"`
		PlayerCount int        `json:"player_count"`
	}

	history := make([]RoomHistory, 0)
	for rows.Next() {
		var rh RoomHistory
		if err := rows.Scan(&rh.ID, &rh.PIN, &rh.QuizID, &rh.QuizTitle,
			&rh.GameMode, &rh.Status, &rh.CreatedAt, &rh.EndedAt, &rh.PlayerCount); err != nil {
			continue
		}
		history = append(history, rh)
	}

	return c.JSON(history)
}

// GET /api/rooms/:pin/info  — public room info (for join screen)
func (h *RoomHandler) Info(c *fiber.Ctx) error {
	pin := c.Params("pin")

	var room struct {
		ID        uuid.UUID `json:"id"`
		Status    string    `json:"status"`
		QuizTitle string    `json:"quiz_title"`
		HostName  string    `json:"host_name"`
	}

	err := h.db.QueryRow(context.Background(),
		`SELECT r.id, r.status, q.title, u.full_name
		 FROM rooms r
		 JOIN quizzes q ON q.id = r.quiz_id
		 JOIN users u ON u.id = r.host_id
		 WHERE r.pin = $1`,
		pin,
	).Scan(&room.ID, &room.Status, &room.QuizTitle, &room.HostName)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Room not found"})
	}

	if room.Status == string(models.RoomCompleted) || room.Status == string(models.RoomAbandoned) {
		return c.Status(410).JSON(fiber.Map{"error": "Game already ended"})
	}

	return c.JSON(room)
}

// WebSocket: ws://host/ws/room/:pin
// Handles both host connection and student join
func (h *RoomHandler) WebSocketHandler(c *fws.Conn) {
	pin := c.Params("pin")
	role := c.Query("role", "student") // "host" or "student"

	client := ws.NewClient(c, h.hub)
	h.hub.Register <- client

	gameRoom, exists := h.hub.GetRoom(pin)

	if role == "host" {
		if !exists {
			gameRoom = h.initializeGameRoom(pin, client)
			if gameRoom == nil {
				client.SendMessage(ws.NewMessage(ws.MsgError, ws.ErrorPayload{Code: "ROOM_NOT_FOUND", Message: "Room not found in database"}))
				return
			}
			h.hub.CreateRoom(gameRoom)
		}
		gameRoom.AddHost(client)
	} else {
		// Students join after room is created
		if !exists {
			client.SendMessage(ws.NewMessage(ws.MsgError, ws.ErrorPayload{Code: "ROOM_NOT_FOUND", Message: "Room not found. Ask your teacher for the PIN."}))
			return
		}
		if gameRoom.Status != models.RoomWaiting {
			client.SendMessage(ws.NewMessage(ws.MsgError, ws.ErrorPayload{Code: "GAME_STARTED", Message: "Game has already started"}))
			return
		}
		// Student will send join_room message with nickname
	}

	go client.WritePump()
	client.ReadPump() // blocks until disconnect
}

func (h *RoomHandler) initializeGameRoom(pin string, hostClient *ws.Client) *ws.GameRoom {
	var roomID, quizID, hostID uuid.UUID
	var gameMode models.GameMode
	var settingsJSON []byte

	err := h.db.QueryRow(context.Background(),
		`SELECT id, quiz_id, host_id, game_mode, settings FROM rooms WHERE pin = $1`,
		pin,
	).Scan(&roomID, &quizID, &hostID, &gameMode, &settingsJSON)
	if err != nil {
		return nil
	}

	var settings models.RoomSettings
	json.Unmarshal(settingsJSON, &settings)

	// Load quiz with questions
	quiz := h.loadFullQuiz(quizID)
	if quiz == nil {
		return nil
	}

	repo := &gameRepo{db: h.db}
	engine := game.NewGameEngine(repo)

	return ws.NewGameRoom(h.hub, roomID, pin, hostID, quiz, gameMode, settings, engine)
}

func (h *RoomHandler) loadFullQuiz(quizID uuid.UUID) *models.Quiz {
	var quiz models.Quiz
	err := h.db.QueryRow(context.Background(),
		`SELECT id, teacher_id, title, total_questions FROM quizzes WHERE id = $1`,
		quizID,
	).Scan(&quiz.ID, &quiz.TeacherID, &quiz.Title, &quiz.TotalQuestions)
	if err != nil {
		return nil
	}

	qh := &QuizHandler{db: h.db}
	questions, err := qh.loadQuestions(quizID)
	if err != nil {
		return nil
	}
	quiz.Questions = questions
	return &quiz
}

// GET /api/rooms/:id/results  — detailed player results for one session
func (h *RoomHandler) Results(c *fiber.Ctx) error {
	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid room id"})
	}
	teacherID := middleware.GetUserID(c)

	// Verify ownership
	var hostID uuid.UUID
	if err := h.db.QueryRow(context.Background(),
		`SELECT host_id FROM rooms WHERE id=$1`, roomID,
	).Scan(&hostID); err != nil || hostID != teacherID {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	type PlayerResult struct {
		Rank          int     `json:"rank"`
		Nickname      string  `json:"nickname"`
		Avatar        string  `json:"avatar"`
		Score         int     `json:"score"`
		Streak        int     `json:"streak"`
		CorrectCount  int     `json:"correct_count"`
		TotalAnswered int     `json:"total_answered"`
		AvgTimeMs     float64 `json:"avg_time_ms"`
	}

	rows, err := h.db.Query(context.Background(),
		`SELECT rp.nickname, rp.avatar, rp.score, rp.streak,
		        COUNT(CASE WHEN ga.is_correct THEN 1 END)::int AS correct_count,
		        COUNT(ga.id)::int AS total_answered,
		        COALESCE(AVG(ga.response_time_ms), 0) AS avg_time_ms
		 FROM room_participants rp
		 LEFT JOIN game_answers ga ON ga.participant_id = rp.id
		 WHERE rp.room_id = $1
		 GROUP BY rp.id
		 ORDER BY rp.score DESC, correct_count DESC`,
		roomID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	results := make([]PlayerResult, 0)
	rank := 1
	for rows.Next() {
		var p PlayerResult
		if err := rows.Scan(&p.Nickname, &p.Avatar, &p.Score, &p.Streak,
			&p.CorrectCount, &p.TotalAnswered, &p.AvgTimeMs); err != nil {
			continue
		}
		p.Rank = rank
		rank++
		results = append(results, p)
	}
	return c.JSON(results)
}

// Hub handles join_room message from student
func (h *RoomHandler) handleStudentJoin(c *ws.Client, pin, nickname, avatar string) error {
	gameRoom, exists := h.hub.GetRoom(pin)
	if !exists {
		return fmt.Errorf("room not found")
	}

	participantID := uuid.New()
	if avatar == "" {
		avatars := []string{"🐶", "🐱", "🐻", "🦊", "🐼", "🐨", "🦁", "🐯"}
		avatar = avatars[rand.Intn(len(avatars))]
	}

	participant := &models.RoomParticipant{
		ID:       participantID,
		RoomID:   gameRoom.RoomID,
		Nickname: nickname,
		Avatar:   avatar,
		IsActive: true,
		JoinedAt: time.Now(),
	}

	// Persist to DB
	_, err := h.db.Exec(context.Background(),
		`INSERT INTO room_participants (id, room_id, nickname, avatar, is_active)
		 VALUES ($1, $2, $3, $4, true)`,
		participantID, gameRoom.RoomID, nickname, avatar,
	)
	if err != nil {
		return fmt.Errorf("failed to save participant: %w", err)
	}

	gameRoom.AddStudent(c, participant)
	return nil
}

func (h *RoomHandler) generatePIN() (string, error) {
	for i := 0; i < 10; i++ {
		pin := fmt.Sprintf("%06d", rand.Intn(1000000))
		var exists bool
		h.db.QueryRow(context.Background(),
			`SELECT EXISTS(SELECT 1 FROM rooms WHERE pin=$1 AND status IN ('waiting','in_progress'))`,
			pin,
		).Scan(&exists)
		if !exists {
			return pin, nil
		}
	}
	return "", fmt.Errorf("could not generate unique PIN")
}

// ─── Game Repository Implementation ─────────────────────────────────────────

type gameRepo struct {
	db *pgxpool.Pool
}

func (r *gameRepo) SaveGameAnswer(ctx context.Context, a *models.GameAnswer) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO game_answers
		 (id, room_id, participant_id, question_id, selected_option_id, text_answer,
		  is_correct, points_earned, response_time_ms, answered_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		uuid.New(), a.RoomID, a.ParticipantID, a.QuestionID, a.SelectedOptionID,
		a.TextAnswer, a.IsCorrect, a.PointsEarned, a.ResponseTimeMs, a.AnsweredAt,
	)
	return err
}

func (r *gameRepo) UpdateParticipantScore(ctx context.Context, participantID uuid.UUID, score, streak int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE room_participants SET score=$1, streak=$2 WHERE id=$3`,
		score, streak, participantID,
	)
	return err
}

func (r *gameRepo) UpdateParticipantTeam(ctx context.Context, participantID uuid.UUID, teamID int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE room_participants SET team_id=$1 WHERE id=$2`,
		teamID, participantID,
	)
	return err
}

func (r *gameRepo) UpdateRoomStatus(ctx context.Context, roomID uuid.UUID, status models.RoomStatus) error {
	extra := ""
	if status == models.RoomInProgress {
		extra = ", started_at=NOW()"
	}
	query := fmt.Sprintf(`UPDATE rooms SET status=$1%s WHERE id=$2`, extra)
	_, err := r.db.Exec(ctx, query, status, roomID)
	return err
}

func (r *gameRepo) FinalizeRoom(ctx context.Context, roomID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE rooms SET status='completed', ended_at=NOW() WHERE id=$1`,
		roomID,
	)
	return err
}
