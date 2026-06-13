package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"gogame.uz/backend/internal/models"
	ws "gogame.uz/backend/internal/websocket"
)

// ParticipantState holds in-memory game state per player
type ParticipantState struct {
	ParticipantID uuid.UUID
	Score         int
	Streak        int
	Rank          int
	RankDelta     int
	Answered      bool
	ConfidenceLevel int
}

// QuestionAnswerState tracks answers for a single question
type QuestionAnswerState struct {
	QuestionID   uuid.UUID
	StartedAt    time.Time
	Answers      map[uuid.UUID]*ws.SubmitAnswerPayload // participantID → answer
	AnswerTimes  map[uuid.UUID]time.Time
	OptionCounts map[string]int
	CorrectCount int
	mu           sync.Mutex
	endOnce      sync.Once // ensures endQuestion runs exactly once per question
}

// GameEngine implements the core game logic state machine
//
// State transitions:
// waiting → in_progress → (per-question loop) → completed
//
// Per-question: show_question → accepting_answers → question_end → leaderboard → next
type GameEngine struct {
	DB     GameRepository
	States map[uuid.UUID]*ParticipantState

	currentQuestion    *QuestionAnswerState
	currentQuestionIdx int
	cancelTimer        context.CancelFunc
	answerLock         sync.Mutex

	// Self-paced / team rejimi holati (answerLock bilan himoyalangan)
	sp *selfPacedState
}

// GameRepository abstracts DB writes (answers, scores)
type GameRepository interface {
	SaveGameAnswer(ctx context.Context, answer *models.GameAnswer) error
	UpdateParticipantScore(ctx context.Context, participantID uuid.UUID, score, streak int) error
	UpdateParticipantTeam(ctx context.Context, participantID uuid.UUID, teamID int) error
	UpdateRoomStatus(ctx context.Context, roomID uuid.UUID, status models.RoomStatus) error
	FinalizeRoom(ctx context.Context, roomID uuid.UUID) error
}

func NewGameEngine(db GameRepository) *GameEngine {
	return &GameEngine{
		DB:     db,
		States: make(map[uuid.UUID]*ParticipantState),
	}
}

// ─── State Machine ───────────────────────────────────────────────────────────

func (e *GameEngine) Start(room *ws.GameRoom) {
	if isSelfPaced(room.GameMode) {
		e.startSelfPaced(room)
		return
	}
	room.Log("Game starting, mode=%s, questions=%d", room.GameMode, len(room.Quiz.Questions))

	// Initialize participant states
	for _, p := range room.GetActiveParticipants() {
		e.States[p.ID] = &ParticipantState{ParticipantID: p.ID}
	}

	room.Status = models.RoomInProgress
	e.DB.UpdateRoomStatus(context.Background(), room.RoomID, models.RoomInProgress)

	room.Broadcast(ws.NewMessage(ws.MsgGameStarted, map[string]any{
		"total_questions": len(room.Quiz.Questions),
		"game_mode":       room.GameMode,
	}))

	// Brief countdown before first question
	time.Sleep(3 * time.Second)
	e.showQuestion(room, 0)
}

func (e *GameEngine) NextQuestion(room *ws.GameRoom) {
	// Self-paced rejimida o'qituvchi savolni boshqarmaydi — har talaba o'zi o'tadi.
	if isSelfPaced(room.GameMode) {
		return
	}
	// Find current index from quiz state
	nextIdx := e.currentQuestionIndex(room) + 1
	if nextIdx >= len(room.Quiz.Questions) {
		e.End(room)
		return
	}
	e.showQuestion(room, nextIdx)
}

func (e *GameEngine) Pause(room *ws.GameRoom) {
	if e.cancelTimer != nil {
		e.cancelTimer()
	}
	room.Status = models.RoomPaused
	room.Broadcast(ws.NewMessage(ws.MsgGamePaused, nil))
}

func (e *GameEngine) Resume(room *ws.GameRoom) {
	room.Status = models.RoomInProgress
	room.Broadcast(ws.NewMessage(ws.MsgGameResumed, nil))
	// Re-show current question with remaining time (simplified: restart timer)
	if e.currentQuestion != nil {
		q := e.getQuestion(room, e.currentQuestionIndexByID(room, e.currentQuestion.QuestionID))
		if q != nil {
			go e.runTimer(room, q, e.currentQuestionIndexByID(room, q.ID))
		}
	}
}

func (e *GameEngine) End(room *ws.GameRoom) {
	if e.cancelTimer != nil {
		e.cancelTimer()
	}
	room.Status = models.RoomCompleted
	e.DB.FinalizeRoom(context.Background(), room.RoomID)

	leaderboard := e.buildLeaderboard(room)
	gameOver := ws.GameOverPayload{
		Leaderboard: leaderboard,
		Stats:       e.buildGameStats(room),
	}
	if room.GameMode == models.ModeTeam {
		gameOver.Teams = e.buildTeamStandings(room)
	}
	room.Broadcast(ws.NewMessage(ws.MsgGameOver, gameOver))
	room.Log("Game ended")
}

// ─── Question Flow ───────────────────────────────────────────────────────────

func (e *GameEngine) showQuestion(room *ws.GameRoom, index int) {
	if index >= len(room.Quiz.Questions) {
		e.End(room)
		return
	}
	q := &room.Quiz.Questions[index]

	e.answerLock.Lock()
	e.currentQuestion = &QuestionAnswerState{
		QuestionID:   q.ID,
		StartedAt:    time.Now(),
		Answers:      make(map[uuid.UUID]*ws.SubmitAnswerPayload),
		AnswerTimes:  make(map[uuid.UUID]time.Time),
		OptionCounts: make(map[string]int),
	}
	e.currentQuestionIdx = index
	// Reset answered state
	for _, s := range e.States {
		s.Answered = false
	}
	e.answerLock.Unlock()

	payload := buildQuestionPayload(q, index, len(room.Quiz.Questions))
	room.Broadcast(ws.NewMessage(ws.MsgQuestion, payload))

	go e.runTimer(room, q, index)
}

func (e *GameEngine) runTimer(room *ws.GameRoom, q *models.Question, index int) {
	ctx, cancel := context.WithCancel(context.Background())
	e.cancelTimer = cancel
	defer cancel()

	// Capture state reference so endOnce belongs to this question
	e.answerLock.Lock()
	state := e.currentQuestion
	e.answerLock.Unlock()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeLeft := q.TimeLimit
	for {
		select {
		case <-ctx.Done():
			// Timer cancelled — auto-end already triggered via HandleAnswer
			return
		case <-ticker.C:
			timeLeft--
			room.Broadcast(ws.NewMessage(ws.MsgTimer, ws.TimerPayload{SecondsLeft: timeLeft}))

			if timeLeft <= 0 {
				if state != nil {
					state.endOnce.Do(func() { e.endQuestion(room, q, index) })
				}
				return
			}
		}
	}
}

func (e *GameEngine) endQuestion(room *ws.GameRoom, q *models.Question, index int) {
	e.answerLock.Lock()
	state := e.currentQuestion
	e.answerLock.Unlock()

	correctIDs := collectCorrectIDs(q)
	stats := ws.QuestionStats{
		TotalAnswers:  len(state.Answers),
		CorrectCount:  state.CorrectCount,
		OptionCounts:  state.OptionCounts,
	}

	// Compute average response time
	if len(state.AnswerTimes) > 0 {
		total := 0
		for _, t := range state.AnswerTimes {
			total += int(t.Sub(state.StartedAt).Milliseconds())
		}
		stats.AverageTimeMs = total / len(state.AnswerTimes)
	}

	var explanation *string
	if q.Explanation != nil {
		explanation = q.Explanation
	}

	room.Broadcast(ws.NewMessage(ws.MsgQuestionEnd, ws.QuestionEndPayload{
		QuestionID:     q.ID.String(),
		CorrectOptions: correctIDs,
		Explanation:    explanation,
		Stats:          stats,
	}))

	// Show leaderboard if enabled
	if room.Settings.ShowLeaderboard {
		time.Sleep(3 * time.Second)
		leaderboard := e.buildLeaderboard(room)
		room.Broadcast(ws.NewMessage(ws.MsgLeaderboard, ws.LeaderboardPayload{
			Players: leaderboard,
		}))
	}
}

// ─── Answer Handling ─────────────────────────────────────────────────────────

// answerOutcome holds all data computed under the lock, so we can act outside it.
type answerOutcome struct {
	result        ws.AnswerResultPayload
	state         *QuestionAnswerState
	question      *models.Question
	answeredCount int
	currentIdx    int
	cancelFn      context.CancelFunc
	logMsg        string
}

func (e *GameEngine) processAnswerLocked(room *ws.GameRoom, participantID uuid.UUID, payload ws.SubmitAnswerPayload) (answerOutcome, bool) {
	e.answerLock.Lock()
	defer e.answerLock.Unlock()

	state := e.currentQuestion
	if state == nil {
		return answerOutcome{}, false
	}

	pState, ok := e.States[participantID]
	if !ok || pState.Answered {
		return answerOutcome{}, false
	}

	qID, err := uuid.Parse(payload.QuestionID)
	if err != nil || qID != state.QuestionID {
		return answerOutcome{}, false
	}

	q := e.findQuestion(room, qID)
	if q == nil {
		return answerOutcome{}, false
	}

	pState.Answered = true
	answerTime := time.Now()
	responseTimeMs := int(answerTime.Sub(state.StartedAt).Milliseconds())
	state.AnswerTimes[participantID] = answerTime

	isCorrect := e.evalAnswer(q, payload)
	answerIndex := len(state.Answers)
	state.Answers[participantID] = &payload
	if payload.OptionID != nil {
		state.OptionCounts[*payload.OptionID]++
	}

	var pointsEarned, speedBonus, streakBonus int
	responseTimeSec := responseTimeMs / 1000

	switch room.GameMode {
	case models.ModeAccuracy:
		if isCorrect {
			pointsEarned = CalcAccuracyPoints(q.Points)
		}
	case models.ModeConfidence:
		pointsEarned = CalcConfidencePoints(q.Points, pState.ConfidenceLevel, isCorrect)
	default:
		if isCorrect {
			pointsEarned, speedBonus, streakBonus = CalcPoints(q.Points, q.TimeLimit, responseTimeSec, pState.Streak, answerIndex)
		}
	}

	if isCorrect {
		pState.Streak++
		state.CorrectCount++
	} else {
		pState.Streak = 0
	}

	if room.GameMode != models.ModeZeroStakes {
		pState.Score += pointsEarned
		go e.DB.UpdateParticipantScore(context.Background(), participantID, pState.Score, pState.Streak)
	}

	gameAnswer := &models.GameAnswer{
		RoomID:         room.RoomID,
		ParticipantID:  participantID,
		QuestionID:     qID,
		IsCorrect:      &isCorrect,
		PointsEarned:   pointsEarned,
		ResponseTimeMs: &responseTimeMs,
		AnsweredAt:     answerTime,
	}
	if payload.OptionID != nil {
		oid, _ := uuid.Parse(*payload.OptionID)
		gameAnswer.SelectedOptionID = &oid
	}
	gameAnswer.TextAnswer = payload.TextAnswer
	go e.DB.SaveGameAnswer(context.Background(), gameAnswer)

	out := answerOutcome{
		result: ws.AnswerResultPayload{
			IsCorrect:    isCorrect,
			PointsEarned: pointsEarned,
			TotalScore:   pState.Score,
			Streak:       pState.Streak,
			StreakBonus:  streakBonus,
			Rank:         pState.Rank,
		},
		state:         state,
		question:      q,
		answeredCount: len(state.Answers),
		currentIdx:    e.currentQuestionIdx,
		cancelFn:      e.cancelTimer,
		logMsg: fmt.Sprintf("[Room %s] %s answered %v (+%d) speed=%ds streak_bonus=%d",
			room.PIN, participantID, isCorrect, speedBonus, responseTimeSec, streakBonus),
	}
	return out, true
}

func (e *GameEngine) HandleAnswer(room *ws.GameRoom, participantID uuid.UUID, payload ws.SubmitAnswerPayload) {
	if isSelfPaced(room.GameMode) {
		e.handleSelfPacedAnswer(room, participantID, payload)
		return
	}
	out, ok := e.processAnswerLocked(room, participantID, payload)
	if !ok {
		return
	}

	log.Print(out.logMsg)

	// Send result to student (outside lock — no deadlock risk)
	room.SendToParticipant(participantID, ws.NewMessage(ws.MsgAnswerResult, out.result))

	// Notify host of live answer count
	activeCount := len(room.GetActiveParticipants())
	room.SendToHost(ws.NewMessage(ws.MsgAnswerCount, ws.AnswerCountPayload{
		Answered: out.answeredCount,
		Total:    activeCount,
	}))

	// Auto-end: all active participants have answered
	if activeCount > 0 && out.answeredCount >= activeCount {
		if out.cancelFn != nil {
			out.cancelFn() // stop timer
		}
		go out.state.endOnce.Do(func() {
			e.endQuestion(room, out.question, out.currentIdx)
		})
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (e *GameEngine) evalAnswer(q *models.Question, payload ws.SubmitAnswerPayload) bool {
	switch q.Type {
	case models.QuestionMultipleChoice, models.QuestionTrueFalse, models.QuestionImageChoice:
		if payload.OptionID == nil {
			return false
		}
		for _, opt := range q.Options {
			if opt.ID.String() == *payload.OptionID {
				return opt.IsCorrect
			}
		}
	case models.QuestionShortAnswer, models.QuestionFillBlank:
		if payload.TextAnswer == nil {
			return false
		}
		for _, opt := range q.Options {
			if opt.IsCorrect && normalize(*payload.TextAnswer) == normalize(opt.OptionText) {
				return true
			}
		}
	case models.QuestionPoll:
		return true // Poll savollarda to'g'ri/noto'g'ri yo'q
	}
	return false
}

func normalize(s string) string {
	// Case-insensitive, trim spaces
	result := []rune{}
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r += 32
		}
		if r != ' ' || (len(result) > 0 && result[len(result)-1] != ' ') {
			result = append(result, r)
		}
	}
	return string(result)
}

func (e *GameEngine) buildLeaderboard(room *ws.GameRoom) []ws.LeaderboardEntry {
	states := make([]*ParticipantState, 0, len(e.States))
	for _, s := range e.States {
		states = append(states, s)
	}
	UpdateLeaderboard(states)

	participants := room.GetActiveParticipants()
	pMap := make(map[uuid.UUID]*models.RoomParticipant)
	for _, p := range participants {
		pMap[p.ID] = p
	}

	entries := make([]ws.LeaderboardEntry, 0, len(states))
	for _, s := range states {
		p, ok := pMap[s.ParticipantID]
		if !ok {
			continue
		}
		entries = append(entries, ws.LeaderboardEntry{
			Rank:     s.Rank,
			ID:       s.ParticipantID.String(),
			Nickname: p.Nickname,
			Avatar:   p.Avatar,
			Score:    s.Score,
			Streak:   s.Streak,
			Delta:    s.RankDelta,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Rank < entries[j].Rank
	})
	return entries
}

func (e *GameEngine) buildGameStats(room *ws.GameRoom) ws.GameStats {
	total := len(e.States)
	correct := 0
	for _, s := range e.States {
		if s.Score > 0 {
			correct++
		}
	}
	acc := 0.0
	if total > 0 {
		acc = float64(correct) / float64(total) * 100
	}
	return ws.GameStats{
		TotalPlayers:    total,
		TotalQuestions:  len(room.Quiz.Questions),
		AverageAccuracy: acc,
	}
}

func collectCorrectIDs(q *models.Question) []string {
	ids := []string{}
	for _, opt := range q.Options {
		if opt.IsCorrect {
			ids = append(ids, opt.ID.String())
		}
	}
	return ids
}

func (e *GameEngine) findQuestion(room *ws.GameRoom, qID uuid.UUID) *models.Question {
	for i := range room.Quiz.Questions {
		if room.Quiz.Questions[i].ID == qID {
			return &room.Quiz.Questions[i]
		}
	}
	return nil
}

func (e *GameEngine) getQuestion(room *ws.GameRoom, index int) *models.Question {
	if index < 0 || index >= len(room.Quiz.Questions) {
		return nil
	}
	return &room.Quiz.Questions[index]
}

func (e *GameEngine) currentQuestionIndex(room *ws.GameRoom) int {
	if e.currentQuestion == nil {
		return -1
	}
	return e.currentQuestionIndexByID(room, e.currentQuestion.QuestionID)
}

func (e *GameEngine) currentQuestionIndexByID(room *ws.GameRoom, qID uuid.UUID) int {
	for i, q := range room.Quiz.Questions {
		if q.ID == qID {
			return i
		}
	}
	return -1
}

func buildQuestionPayload(q *models.Question, index, total int) ws.QuestionPayload {
	opts := make([]ws.StudentOptionPayload, len(q.Options))
	for i, o := range q.Options {
		opts[i] = ws.StudentOptionPayload{
			ID:         o.ID.String(),
			OptionText: o.OptionText,
			MediaURL:   o.MediaURL,
		}
	}
	return ws.QuestionPayload{
		QuestionIndex:  index,
		TotalQuestions: total,
		QuestionID:     q.ID.String(),
		Type:           string(q.Type),
		QuestionText:   q.QuestionText,
		MediaURL:       q.MediaURL,
		MediaType:      q.MediaType,
		TimeLimit:      q.TimeLimit,
		Points:         q.Points,
		Options:        opts,
	}
}

// Ensure GameEngine implements GameEngineInterface
var _ ws.GameEngineInterface = (*GameEngine)(nil)

// JSON marshal helper to send raw
func toJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
