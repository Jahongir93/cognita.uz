package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RoomStatus string

const (
	RoomWaiting    RoomStatus = "waiting"
	RoomInProgress RoomStatus = "in_progress"
	RoomPaused     RoomStatus = "paused"
	RoomCompleted  RoomStatus = "completed"
	RoomAbandoned  RoomStatus = "abandoned"
)

type GameMode string

const (
	ModeClassic    GameMode = "classic"
	ModeTeam       GameMode = "team"
	ModeAccuracy   GameMode = "accuracy"
	ModeConfidence GameMode = "confidence"
	ModeZeroStakes GameMode = "zero_stakes"
)

type RoomSettings struct {
	ShuffleQuestions bool `json:"shuffle_questions"`
	ShuffleAnswers   bool `json:"shuffle_answers"`
	ShowLeaderboard  bool `json:"show_leaderboard"`
	Music            bool `json:"music"`
	LobbyMusic       bool `json:"lobby_music"`
	TeamCount        int  `json:"team_count"`
	ShowCorrectAnswer bool `json:"show_correct_answer"`
}

func DefaultRoomSettings() RoomSettings {
	return RoomSettings{
		ShuffleQuestions:  false,
		ShuffleAnswers:    false,
		ShowLeaderboard:   true,
		Music:             true,
		LobbyMusic:        true,
		TeamCount:         0,
		ShowCorrectAnswer: true,
	}
}

type Room struct {
	ID                    uuid.UUID      `json:"id" db:"id"`
	QuizID                uuid.UUID      `json:"quiz_id" db:"quiz_id"`
	HostID                uuid.UUID      `json:"host_id" db:"host_id"`
	PIN                   string         `json:"pin" db:"pin"`
	Status                RoomStatus     `json:"status" db:"status"`
	GameMode              GameMode       `json:"game_mode" db:"game_mode"`
	Settings              json.RawMessage `json:"settings" db:"settings"`
	CurrentQuestionIndex  int            `json:"current_question_index" db:"current_question_index"`
	StartedAt             *time.Time     `json:"started_at" db:"started_at"`
	EndedAt               *time.Time     `json:"ended_at" db:"ended_at"`
	CreatedAt             time.Time      `json:"created_at" db:"created_at"`

	Quiz         *Quiz              `json:"quiz,omitempty"`
	Participants []RoomParticipant  `json:"participants,omitempty"`
}

type RoomParticipant struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	RoomID   uuid.UUID  `json:"room_id" db:"room_id"`
	UserID   *uuid.UUID `json:"user_id" db:"user_id"`
	Nickname string     `json:"nickname" db:"nickname"`
	Avatar   string     `json:"avatar" db:"avatar"`
	TeamID   *int       `json:"team_id" db:"team_id"`
	Score    int        `json:"score" db:"score"`
	Streak   int        `json:"streak" db:"streak"`
	Rank     *int       `json:"rank" db:"rank"`
	IsActive bool       `json:"is_active" db:"is_active"`
	JoinedAt time.Time  `json:"joined_at" db:"joined_at"`
}

type GameAnswer struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	RoomID          uuid.UUID  `json:"room_id" db:"room_id"`
	ParticipantID   uuid.UUID  `json:"participant_id" db:"participant_id"`
	QuestionID      uuid.UUID  `json:"question_id" db:"question_id"`
	SelectedOptionID *uuid.UUID `json:"selected_option_id" db:"selected_option_id"`
	TextAnswer      *string    `json:"text_answer" db:"text_answer"`
	IsCorrect       *bool      `json:"is_correct" db:"is_correct"`
	PointsEarned    int        `json:"points_earned" db:"points_earned"`
	ResponseTimeMs  *int       `json:"response_time_ms" db:"response_time_ms"`
	AnsweredAt      time.Time  `json:"answered_at" db:"answered_at"`
}

type CreateRoomRequest struct {
	QuizID   uuid.UUID    `json:"quiz_id" validate:"required"`
	GameMode GameMode     `json:"game_mode"`
	Settings RoomSettings `json:"settings"`
}

type JoinRoomRequest struct {
	PIN      string `json:"pin" validate:"required"`
	Nickname string `json:"nickname" validate:"required,min=2,max=30"`
	Avatar   string `json:"avatar"`
}

// ─── Assignment models ──────────────────────────────────────────────────────

type AssignmentStatus string

const (
	AssignmentDraft  AssignmentStatus = "draft"
	AssignmentActive AssignmentStatus = "active"
	AssignmentClosed AssignmentStatus = "closed"
)

type AssignmentSettings struct {
	AllowRetries      bool `json:"allow_retries"`
	MaxRetries        int  `json:"max_retries"`
	ShowAnswers       bool `json:"show_answers"`
	TimerEnabled      bool `json:"timer_enabled"`
	ReadAloud         bool `json:"read_aloud"`
	FocusMode         bool `json:"focus_mode"`
}

type Assignment struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	QuizID    uuid.UUID        `json:"quiz_id" db:"quiz_id"`
	TeacherID uuid.UUID        `json:"teacher_id" db:"teacher_id"`
	ClassID   *uuid.UUID       `json:"class_id" db:"class_id"`
	Title     string           `json:"title" db:"title"`
	Status    AssignmentStatus `json:"status" db:"status"`
	DueDate   *time.Time       `json:"due_date" db:"due_date"`
	Settings  json.RawMessage  `json:"settings" db:"settings"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}
