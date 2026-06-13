package game

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"gogame.uz/backend/internal/models"
	ws "gogame.uz/backend/internal/websocket"
)

// Self-paced ("Mustaqil") va team ("Jamoaviy") rejimlari uchun mantiq.
//
// Klassik rejimdan farqi: o'qituvchi savolni boshqarmaydi. Har bir talaba
// o'z qurilmasida, tasodifiy tartibda savollarni o'z tezligida yechadi.
// Javob bergach darrov keyingi savol shu talabaga yuboriladi. O'qituvchi
// faqat jonli progress va yakuniy natijalarni ko'radi. Vaqt hisobi mijoz
// tomonida yuritiladi (har savol time_limit), vaqt tugasa mijoz bo'sh javob
// yuboradi.

func isSelfPaced(m models.GameMode) bool {
	return m == models.ModeSelfPaced || m == models.ModeTeam
}

// selfPacedState — answerLock bilan himoyalanadi.
type selfPacedState struct {
	queues    map[uuid.UUID][]int // participant → aralashtirilgan savol indekslari
	pos       map[uuid.UUID]int   // participant → keyingi savol o'rni
	finished  map[uuid.UUID]bool
	teamOf    map[uuid.UUID]int   // participant → jamoa raqami (faqat team rejimi)
	teamNames map[int]string
}

func (e *GameEngine) startSelfPaced(room *ws.GameRoom) {
	room.Log("Self-paced game starting, mode=%s, questions=%d", room.GameMode, len(room.Quiz.Questions))

	participants := room.GetActiveParticipants()
	nq := len(room.Quiz.Questions)

	sp := &selfPacedState{
		queues:    make(map[uuid.UUID][]int),
		pos:       make(map[uuid.UUID]int),
		finished:  make(map[uuid.UUID]bool),
		teamOf:    make(map[uuid.UUID]int),
		teamNames: make(map[int]string),
	}

	// Jamoa rejimida jamoalar sonini aniqlash
	teamCount := 0
	if room.GameMode == models.ModeTeam {
		teamCount = room.Settings.TeamCount
		if teamCount < 2 {
			teamCount = 2
		}
		for i := 1; i <= teamCount; i++ {
			sp.teamNames[i] = fmt.Sprintf("Jamoa %d", i)
		}
	}

	e.answerLock.Lock()
	for idx, p := range participants {
		e.States[p.ID] = &ParticipantState{ParticipantID: p.ID}

		// Tasodifiy savol tartibi
		order := make([]int, nq)
		for i := range order {
			order[i] = i
		}
		rand.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		sp.queues[p.ID] = order
		sp.pos[p.ID] = 0
		sp.finished[p.ID] = false

		// Jamoaga avtomatik (round-robin) taqsimlash
		if teamCount > 0 {
			team := (idx % teamCount) + 1
			sp.teamOf[p.ID] = team
			t := team
			p.TeamID = &t
			go e.DB.UpdateParticipantTeam(context.Background(), p.ID, t)
		}
	}
	e.sp = sp
	e.answerLock.Unlock()

	room.Status = models.RoomInProgress
	e.DB.UpdateRoomStatus(context.Background(), room.RoomID, models.RoomInProgress)

	room.Broadcast(ws.NewMessage(ws.MsgGameStarted, map[string]any{
		"total_questions": nq,
		"game_mode":       string(room.GameMode),
	}))

	// Birinchi savolni har bir talabaga personal yuborish
	for _, p := range participants {
		e.sendSelfPacedQuestion(room, p.ID)
	}
	e.broadcastProgress(room)
}

// sendSelfPacedQuestion talabaning navbatdagi savolini yuboradi yoki
// savollar tugagan bo'lsa shaxsiy "game_over" yuboradi.
func (e *GameEngine) sendSelfPacedQuestion(room *ws.GameRoom, pid uuid.UUID) {
	e.answerLock.Lock()
	order := e.sp.queues[pid]
	pos := e.sp.pos[pid]
	e.answerLock.Unlock()

	if pos >= len(order) {
		e.answerLock.Lock()
		e.sp.finished[pid] = true
		e.answerLock.Unlock()

		room.SendToParticipant(pid, ws.NewMessage(ws.MsgGameOver, e.buildSelfPacedGameOver(room)))
		e.checkAllFinished(room)
		return
	}

	q := &room.Quiz.Questions[order[pos]]
	payload := buildQuestionPayload(q, pos, len(order))
	room.SendToParticipant(pid, ws.NewMessage(ws.MsgQuestion, payload))
}

func (e *GameEngine) handleSelfPacedAnswer(room *ws.GameRoom, pid uuid.UUID, payload ws.SubmitAnswerPayload) {
	qID, err := uuid.Parse(payload.QuestionID)
	if err != nil {
		return
	}
	q := e.findQuestion(room, qID)
	if q == nil {
		return
	}

	isCorrect := e.evalAnswer(q, payload)
	responseTimeMs := payload.ResponseTimeMs
	if responseTimeMs < 0 {
		responseTimeMs = 0
	}

	e.answerLock.Lock()
	pState, ok := e.States[pid]
	if !ok {
		e.answerLock.Unlock()
		return
	}
	// Bir savolga ikki marta javobni oldini olish: faqat hozirgi savolga javob qabul qilamiz
	curPos := e.sp.pos[pid]
	order := e.sp.queues[pid]
	if curPos >= len(order) || room.Quiz.Questions[order[curPos]].ID != qID {
		e.answerLock.Unlock()
		return
	}

	var pointsEarned, streakBonus int
	if isCorrect {
		pointsEarned, _, streakBonus = CalcPoints(q.Points, q.TimeLimit, responseTimeMs/1000, pState.Streak, 0)
		pState.Streak++
	} else {
		pState.Streak = 0
	}
	pState.Score += pointsEarned
	e.sp.pos[pid]++
	newPos := e.sp.pos[pid]
	score := pState.Score
	streak := pState.Streak
	e.answerLock.Unlock()

	go e.DB.UpdateParticipantScore(context.Background(), pid, score, streak)

	gameAnswer := &models.GameAnswer{
		RoomID:         room.RoomID,
		ParticipantID:  pid,
		QuestionID:     qID,
		IsCorrect:      &isCorrect,
		PointsEarned:   pointsEarned,
		ResponseTimeMs: &responseTimeMs,
		AnsweredAt:     time.Now(),
	}
	if payload.OptionID != nil {
		if oid, e2 := uuid.Parse(*payload.OptionID); e2 == nil {
			gameAnswer.SelectedOptionID = &oid
		}
	}
	gameAnswer.TextAnswer = payload.TextAnswer
	go e.DB.SaveGameAnswer(context.Background(), gameAnswer)

	// Talabaga javob natijasi
	room.SendToParticipant(pid, ws.NewMessage(ws.MsgAnswerResult, ws.AnswerResultPayload{
		IsCorrect:    isCorrect,
		PointsEarned: pointsEarned,
		TotalScore:   score,
		Streak:       streak,
		StreakBonus:  streakBonus,
	}))

	// Keyingi savol yoki tugatish
	if newPos >= len(order) {
		e.answerLock.Lock()
		e.sp.finished[pid] = true
		e.answerLock.Unlock()
		room.SendToParticipant(pid, ws.NewMessage(ws.MsgGameOver, e.buildSelfPacedGameOver(room)))
		e.broadcastProgress(room)
		e.checkAllFinished(room)
		return
	}

	e.broadcastProgress(room)

	// Talaba javob natijasini ko'rishi uchun qisqa pauza, so'ng keyingi savol.
	time.Sleep(1500 * time.Millisecond)
	nextQ := &room.Quiz.Questions[order[newPos]]
	room.SendToParticipant(pid, ws.NewMessage(ws.MsgQuestion, buildQuestionPayload(nextQ, newPos, len(order))))
}

// broadcastProgress hostga har talabaning jonli progressini yuboradi.
func (e *GameEngine) broadcastProgress(room *ws.GameRoom) {
	participants := room.GetActiveParticipants()
	pMap := make(map[uuid.UUID]*models.RoomParticipant, len(participants))
	for _, p := range participants {
		pMap[p.ID] = p
	}

	e.answerLock.Lock()
	players := make([]ws.SelfPacedPlayerProgress, 0, len(participants))
	for _, p := range participants {
		st := e.States[p.ID]
		score := 0
		if st != nil {
			score = st.Score
		}
		var teamID *int
		if t, ok := e.sp.teamOf[p.ID]; ok && t > 0 {
			tv := t
			teamID = &tv
		}
		players = append(players, ws.SelfPacedPlayerProgress{
			ID:       p.ID.String(),
			Nickname: p.Nickname,
			Avatar:   p.Avatar,
			TeamID:   teamID,
			Answered: e.sp.pos[p.ID],
			Total:    len(e.sp.queues[p.ID]),
			Score:    score,
			Finished: e.sp.finished[p.ID],
		})
	}
	e.answerLock.Unlock()

	sort.Slice(players, func(i, j int) bool { return players[i].Score > players[j].Score })

	payload := ws.SelfPacedProgressPayload{Players: players}
	if room.GameMode == models.ModeTeam {
		payload.Teams = e.buildTeamStandings(room)
	}
	room.SendToHost(ws.NewMessage(ws.MsgSelfPacedProgress, payload))
}

// checkAllFinished — barcha faol talabalar tugatgan bo'lsa, o'yinni yakunlaydi.
func (e *GameEngine) checkAllFinished(room *ws.GameRoom) {
	participants := room.GetActiveParticipants()
	if len(participants) == 0 {
		return
	}
	e.answerLock.Lock()
	allDone := true
	for _, p := range participants {
		if !e.sp.finished[p.ID] {
			allDone = false
			break
		}
	}
	e.answerLock.Unlock()
	if allDone {
		e.End(room)
	}
}

// buildSelfPacedGameOver — talabaga yuboriladigan yakuniy natija.
func (e *GameEngine) buildSelfPacedGameOver(room *ws.GameRoom) ws.GameOverPayload {
	out := ws.GameOverPayload{
		Leaderboard: e.buildLeaderboard(room),
		Stats:       e.buildGameStats(room),
	}
	if room.GameMode == models.ModeTeam {
		out.Teams = e.buildTeamStandings(room)
	}
	return out
}

// buildTeamStandings jamoalar ballarini yig'ib, reyting bo'yicha qaytaradi.
func (e *GameEngine) buildTeamStandings(room *ws.GameRoom) []ws.TeamStanding {
	e.answerLock.Lock()
	defer e.answerLock.Unlock()

	if e.sp == nil {
		return nil
	}

	totals := make(map[int]int)
	members := make(map[int]int)
	for pid, team := range e.sp.teamOf {
		if team == 0 {
			continue
		}
		members[team]++
		if st, ok := e.States[pid]; ok {
			totals[team] += st.Score
		}
	}

	standings := make([]ws.TeamStanding, 0, len(e.sp.teamNames))
	for team, name := range e.sp.teamNames {
		standings = append(standings, ws.TeamStanding{
			TeamID:  team,
			Name:    name,
			Score:   totals[team],
			Members: members[team],
		})
	}
	sort.Slice(standings, func(i, j int) bool { return standings[i].Score > standings[j].Score })
	for i := range standings {
		standings[i].Rank = i + 1
	}
	return standings
}
