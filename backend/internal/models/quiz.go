package models

import (
	"time"

	"github.com/google/uuid"
)

type QuestionType string

const (
	QuestionMultipleChoice QuestionType = "multiple_choice"
	QuestionTrueFalse      QuestionType = "true_false"
	QuestionShortAnswer    QuestionType = "short_answer"
	QuestionPoll           QuestionType = "poll"
	QuestionImageChoice    QuestionType = "image_choice"
	QuestionAudio          QuestionType = "audio"
	QuestionDraw           QuestionType = "draw"
	QuestionReorder        QuestionType = "reorder"
	QuestionFillBlank      QuestionType = "fill_blank"
)

type ActivityTemplate string

const (
	TemplateQuiz           ActivityTemplate = "quiz"
	TemplateMatchUp        ActivityTemplate = "match_up"
	TemplateWordsearch     ActivityTemplate = "wordsearch"
	TemplateCrossword      ActivityTemplate = "crossword"
	TemplateMazeChase      ActivityTemplate = "maze_chase"
	TemplateGroupSort      ActivityTemplate = "group_sort"
	TemplateTypeAnswer     ActivityTemplate = "type_answer"
	TemplateLabelDiagram   ActivityTemplate = "labelled_diagram"
	TemplateFlashcards     ActivityTemplate = "flashcards"
)

type Quiz struct {
	ID             uuid.UUID        `json:"id" db:"id"`
	TeacherID      uuid.UUID        `json:"teacher_id" db:"teacher_id"`
	Title          string           `json:"title" db:"title"`
	Description    *string          `json:"description" db:"description"`
	Subject        *string          `json:"subject" db:"subject"`
	GradeLevel     *string          `json:"grade_level" db:"grade_level"`
	CoverImage     *string          `json:"cover_image" db:"cover_image"`
	Template       ActivityTemplate `json:"template" db:"template"`
	IsPublic       bool             `json:"is_public" db:"is_public"`
	TotalQuestions int              `json:"total_questions" db:"total_questions"`
	PlayCount      int              `json:"play_count" db:"play_count"`
	Tags           []string         `json:"tags" db:"tags"`
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at" db:"updated_at"`

	Questions []Question `json:"questions,omitempty"`
	Teacher   *UserPublic `json:"teacher,omitempty"`
}

type Question struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	QuizID       uuid.UUID    `json:"quiz_id" db:"quiz_id"`
	OrderIndex   int          `json:"order_index" db:"order_index"`
	Type         QuestionType `json:"type" db:"type"`
	QuestionText string       `json:"question_text" db:"question_text"`
	MediaURL     *string      `json:"media_url" db:"media_url"`
	MediaType    *string      `json:"media_type" db:"media_type"`
	TimeLimit    int          `json:"time_limit" db:"time_limit"`
	Points       int          `json:"points" db:"points"`
	Explanation  *string      `json:"explanation" db:"explanation"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`

	Options []AnswerOption `json:"options,omitempty"`
}

type AnswerOption struct {
	ID         uuid.UUID `json:"id" db:"id"`
	QuestionID uuid.UUID `json:"question_id" db:"question_id"`
	OptionText string    `json:"option_text" db:"option_text"`
	MediaURL   *string   `json:"media_url" db:"media_url"`
	IsCorrect  bool      `json:"is_correct" db:"is_correct"`
	OrderIndex int       `json:"order_index" db:"order_index"`
}

// StudentAnswerOption hides IsCorrect from students
type StudentAnswerOption struct {
	ID         uuid.UUID `json:"id"`
	OptionText string    `json:"option_text"`
	MediaURL   *string   `json:"media_url"`
	OrderIndex int       `json:"order_index"`
}

func (a *AnswerOption) ToStudent() StudentAnswerOption {
	return StudentAnswerOption{
		ID:         a.ID,
		OptionText: a.OptionText,
		MediaURL:   a.MediaURL,
		OrderIndex: a.OrderIndex,
	}
}

type CreateQuizRequest struct {
	Title       string           `json:"title" validate:"required,min=3"`
	Description *string          `json:"description"`
	Subject     *string          `json:"subject"`
	GradeLevel  *string          `json:"grade_level"`
	Template    ActivityTemplate `json:"template"`
	IsPublic    bool             `json:"is_public"`
	Tags        []string         `json:"tags"`
	Questions   []CreateQuestionRequest `json:"questions"`
}

type CreateQuestionRequest struct {
	OrderIndex   int          `json:"order_index"`
	Type         QuestionType `json:"type" validate:"required"`
	QuestionText string       `json:"question_text" validate:"required"`
	MediaURL     *string      `json:"media_url"`
	MediaType    *string      `json:"media_type"`
	TimeLimit    int          `json:"time_limit"`
	Points       int          `json:"points"`
	Explanation  *string      `json:"explanation"`
	Options      []CreateOptionRequest `json:"options"`
}

type CreateOptionRequest struct {
	OptionText string  `json:"option_text" validate:"required"`
	MediaURL   *string `json:"media_url"`
	IsCorrect  bool    `json:"is_correct"`
	OrderIndex int     `json:"order_index"`
}
