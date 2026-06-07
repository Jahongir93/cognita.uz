package websocket

import "encoding/json"

// ─── Message Types ──────────────────────────────────────────────────────────
//
// C→S: client to server
// S→C: server to client (broadcast yoki personal)

// Client → Server
const (
	MsgJoinRoom       = "join_room"       // student: xonaga qo'shilish
	MsgSubmitAnswer   = "submit_answer"   // student: javob yuborish
	MsgConfidencePick = "confidence_pick" // student: confidence mode uchun
	MsgEmoji          = "send_emoji"      // student: emoji reaction

	// Host only
	MsgStartGame    = "start_game"    // o'yinni boshlash
	MsgNextQuestion = "next_question" // keyingi savol
	MsgPauseGame    = "pause_game"    // pauza
	MsgResumeGame   = "resume_game"   // davom ettirish
	MsgEndGame      = "end_game"      // o'yinni yakunlash
	MsgKickPlayer   = "kick_player"   // o'yinchini chiqarish
)

// Server → Client
const (
	MsgRoomState      = "room_state"      // xona holati (join qilganda)
	MsgPlayerJoined   = "player_joined"   // yangi o'yinchi
	MsgPlayerLeft     = "player_left"     // o'yinchi chiqdi
	MsgGameStarted    = "game_started"    // o'yin boshlandi
	MsgQuestion       = "question"        // savol ko'rsatish
	MsgQuestionEnd    = "question_end"    // savol vaqti tugadi
	MsgAnswerResult   = "answer_result"   // o'yinchi uchun javob natijasi
	MsgAnswerCount    = "answer_count"    // host: nechta o'yinchi javob berdi (live)
	MsgLeaderboard    = "leaderboard"     // liderlar jadvali
	MsgGamePaused     = "game_paused"     // pauza
	MsgGameResumed    = "game_resumed"    // davom etdi
	MsgGameOver       = "game_over"       // o'yin tugadi
	MsgTimer          = "timer"           // real-time timer tick
	MsgEmojiReaction  = "emoji_reaction"  // barcha ko'radigan emoji
	MsgYourInfo       = "your_info"       // student: o'z participant_id si
	MsgError          = "error"           // xato xabari
)

// ─── Base Message ────────────────────────────────────────────────────────────

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func NewMessage(msgType string, payload any) Message {
	data, _ := json.Marshal(payload)
	return Message{Type: msgType, Payload: data}
}

// ─── Payloads: Client → Server ───────────────────────────────────────────────

type JoinRoomPayload struct {
	PIN      string `json:"pin"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Token    string `json:"token,omitempty"` // optional: logged-in user
}

type SubmitAnswerPayload struct {
	QuestionID     string  `json:"question_id"`
	OptionID       *string `json:"option_id,omitempty"`
	TextAnswer     *string `json:"text_answer,omitempty"`
	ResponseTimeMs int     `json:"response_time_ms"`
}

type ConfidencePayload struct {
	QuestionID string `json:"question_id"`
	Level      int    `json:"level"` // 1=unsure, 2=maybe, 3=sure
}

type EmojiPayload struct {
	Emoji string `json:"emoji"`
}

type KickPlayerPayload struct {
	ParticipantID string `json:"participant_id"`
}

// ─── Payloads: Server → Client ───────────────────────────────────────────────

type RoomStatePayload struct {
	RoomID      string                `json:"room_id"`
	PIN         string                `json:"pin"`
	Status      string                `json:"status"`
	GameMode    string                `json:"game_mode"`
	QuizTitle   string                `json:"quiz_title"`
	HostName    string                `json:"host_name"`
	Players     []PlayerInfo          `json:"players"`
	TotalQuestions int               `json:"total_questions"`
}

type PlayerInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Score    int    `json:"score"`
	Streak   int    `json:"streak"`
	TeamID   *int   `json:"team_id,omitempty"`
	IsActive bool   `json:"is_active"`
}

type PlayerJoinedPayload struct {
	Player     PlayerInfo `json:"player"`
	TotalCount int        `json:"total_count"`
}

type QuestionPayload struct {
	QuestionIndex  int                    `json:"question_index"`
	TotalQuestions int                    `json:"total_questions"`
	QuestionID     string                 `json:"question_id"`
	Type           string                 `json:"type"`
	QuestionText   string                 `json:"question_text"`
	MediaURL       *string                `json:"media_url,omitempty"`
	MediaType      *string                `json:"media_type,omitempty"`
	TimeLimit      int                    `json:"time_limit"`
	Points         int                    `json:"points"`
	Options        []StudentOptionPayload `json:"options"`
}

type StudentOptionPayload struct {
	ID         string  `json:"id"`
	OptionText string  `json:"option_text"`
	MediaURL   *string `json:"media_url,omitempty"`
}

type QuestionEndPayload struct {
	QuestionID     string              `json:"question_id"`
	CorrectOptions []string            `json:"correct_options"`
	Explanation    *string             `json:"explanation,omitempty"`
	Stats          QuestionStats       `json:"stats"`
}

type QuestionStats struct {
	TotalAnswers   int            `json:"total_answers"`
	CorrectCount   int            `json:"correct_count"`
	OptionCounts   map[string]int `json:"option_counts"`
	AverageTimeMs  int            `json:"average_time_ms"`
}

type AnswerResultPayload struct {
	IsCorrect    bool   `json:"is_correct"`
	PointsEarned int    `json:"points_earned"`
	TotalScore   int    `json:"total_score"`
	Streak       int    `json:"streak"`
	Rank         int    `json:"rank"`
	StreakBonus  int    `json:"streak_bonus"`
}

type LeaderboardPayload struct {
	Players []LeaderboardEntry `json:"players"`
}

type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Score    int    `json:"score"`
	Streak   int    `json:"streak"`
	Delta    int    `json:"delta"` // rank change from previous question
}

type GameOverPayload struct {
	Leaderboard []LeaderboardEntry `json:"leaderboard"`
	Stats       GameStats          `json:"stats"`
}

type GameStats struct {
	TotalPlayers    int     `json:"total_players"`
	TotalQuestions  int     `json:"total_questions"`
	AverageAccuracy float64 `json:"average_accuracy"`
	DurationSec     int     `json:"duration_sec"`
}

type TimerPayload struct {
	SecondsLeft int `json:"seconds_left"`
}

type AnswerCountPayload struct {
	Answered int `json:"answered"`
	Total    int `json:"total"`
}

type YourInfoPayload struct {
	ParticipantID string `json:"participant_id"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
