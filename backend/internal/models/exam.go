package models

import (
	"time"

	"github.com/google/uuid"
)

type ExamStatus string

const (
	ExamDraft  ExamStatus = "draft"
	ExamActive ExamStatus = "active"
	ExamClosed ExamStatus = "closed"
)

type Exam struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	QuizID      uuid.UUID  `json:"quiz_id" db:"quiz_id"`
	TeacherID   uuid.UUID  `json:"teacher_id" db:"teacher_id"`
	Title       string     `json:"title" db:"title"`
	Code        string     `json:"code" db:"code"`
	TimeLimit   int        `json:"time_limit" db:"time_limit"` // minutes
	StartDate   *time.Time `json:"start_date" db:"start_date"`
	EndDate     *time.Time `json:"end_date" db:"end_date"`
	Shuffle     bool       `json:"shuffle" db:"shuffle"`
	MaxAttempts int        `json:"max_attempts" db:"max_attempts"`
	Status      ExamStatus `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	// computed
	QuizTitle       string `json:"quiz_title,omitempty"`
	SubmissionCount int    `json:"submission_count,omitempty"`
}

type ExamSubmission struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	ExamID      uuid.UUID         `json:"exam_id" db:"exam_id"`
	StudentName string            `json:"student_name" db:"student_name"`
	Answers     []SubmittedAnswer `json:"answers"`
	Score       int               `json:"score" db:"score"`
	MaxScore    int               `json:"max_score" db:"max_score"`
	TimeTaken   int               `json:"time_taken" db:"time_taken"` // seconds
	SubmittedAt time.Time         `json:"submitted_at" db:"submitted_at"`
}

type SubmittedAnswer struct {
	QuestionID string `json:"question_id"`
	OptionID   string `json:"option_id,omitempty"`
	TextAnswer string `json:"text_answer,omitempty"`
}

type CreateExamRequest struct {
	QuizID      string     `json:"quiz_id"`
	Title       string     `json:"title"`
	TimeLimit   int        `json:"time_limit"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Shuffle     bool       `json:"shuffle"`
	MaxAttempts int        `json:"max_attempts"`
}

// ─── Olympiad ────────────────────────────────────────────────────────────────

type OlympiadStatus string

const (
	OlympiadUpcoming  OlympiadStatus = "upcoming"
	OlympiadActive    OlympiadStatus = "active"
	OlympiadCompleted OlympiadStatus = "completed"
)

type Olympiad struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	QuizID          uuid.UUID      `json:"quiz_id" db:"quiz_id"`
	TeacherID       uuid.UUID      `json:"teacher_id" db:"teacher_id"`
	Title           string         `json:"title" db:"title"`
	Description     string         `json:"description" db:"description"`
	Code            string         `json:"code" db:"code"`
	StartTime       time.Time      `json:"start_time" db:"start_time"`
	EndTime         time.Time      `json:"end_time" db:"end_time"`
	MaxParticipants *int           `json:"max_participants" db:"max_participants"`
	Status          OlympiadStatus `json:"status" db:"status"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	// computed
	QuizTitle        string `json:"quiz_title,omitempty"`
	ParticipantCount int    `json:"participant_count,omitempty"`
}

type OlympiadSubmission struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	OlympiadID  uuid.UUID         `json:"olympiad_id" db:"olympiad_id"`
	StudentName string            `json:"student_name" db:"student_name"`
	Answers     []SubmittedAnswer `json:"answers"`
	Score       int               `json:"score" db:"score"`
	MaxScore    int               `json:"max_score" db:"max_score"`
	TimeTaken   int               `json:"time_taken" db:"time_taken"`
	Rank        int               `json:"rank,omitempty"`
	SubmittedAt time.Time         `json:"submitted_at" db:"submitted_at"`
}

type CreateOlympiadRequest struct {
	QuizID          string    `json:"quiz_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	MaxParticipants *int      `json:"max_participants"`
}
