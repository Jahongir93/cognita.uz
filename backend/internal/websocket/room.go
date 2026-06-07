package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"gogame.uz/backend/internal/models"
)

// GameRoom manages the live state of a single game session
type GameRoom struct {
	PIN    string
	RoomID uuid.UUID
	HostID uuid.UUID
	Quiz   *models.Quiz

	// All clients: clientID → *Client
	Clients map[string]*Client

	// Participants: participantID → *models.RoomParticipant
	Participants map[uuid.UUID]*models.RoomParticipant

	// clientID → participantID mapping
	ClientParticipant map[string]uuid.UUID

	GameMode models.GameMode
	Settings models.RoomSettings
	Status   models.RoomStatus

	Hub    *Hub
	Engine GameEngineInterface

	mu sync.RWMutex
}

type GameEngineInterface interface {
	Start(room *GameRoom)
	NextQuestion(room *GameRoom)
	Pause(room *GameRoom)
	Resume(room *GameRoom)
	End(room *GameRoom)
	HandleAnswer(room *GameRoom, participantID uuid.UUID, payload SubmitAnswerPayload)
}

func NewGameRoom(hub *Hub, roomID uuid.UUID, pin string, hostID uuid.UUID, quiz *models.Quiz,
	mode models.GameMode, settings models.RoomSettings, engine GameEngineInterface) *GameRoom {

	return &GameRoom{
		PIN:               pin,
		RoomID:            roomID,
		HostID:            hostID,
		Quiz:              quiz,
		Clients:           make(map[string]*Client),
		Participants:      make(map[uuid.UUID]*models.RoomParticipant),
		ClientParticipant: make(map[string]uuid.UUID),
		GameMode:          mode,
		Settings:          settings,
		Status:            models.RoomWaiting,
		Hub:               hub,
		Engine:            engine,
	}
}

// AddHost registers the host client
func (r *GameRoom) AddHost(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c.Room = r
	c.Role = ClientHost
	r.Clients[c.ID] = c

	// Send current room state to host
	c.SendMessage(NewMessage(MsgRoomState, r.buildRoomState()))
}

// AddStudent adds a student participant
func (r *GameRoom) AddStudent(c *Client, participant *models.RoomParticipant) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c.Room = r
	c.Role = ClientStudent
	c.ParticipantID = participant.ID
	c.Nickname = participant.Nickname
	c.Avatar = participant.Avatar
	r.Clients[c.ID] = c
	r.Participants[participant.ID] = participant
	r.ClientParticipant[c.ID] = participant.ID

	// Notify everyone about new player
	info := playerToInfo(participant)
	r.broadcastLocked(NewMessage(MsgPlayerJoined, PlayerJoinedPayload{
		Player:     info,
		TotalCount: len(r.Participants),
	}))

	// Send room state to the new student
	c.SendMessage(NewMessage(MsgRoomState, r.buildRoomState()))
	// Tell the student their own participant ID
	c.SendMessage(NewMessage(MsgYourInfo, YourInfoPayload{
		ParticipantID: participant.ID.String(),
		Nickname:      participant.Nickname,
		Avatar:        participant.Avatar,
	}))
}

func (r *GameRoom) handleDisconnect(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Clients, c.ID)

	if c.Role == ClientStudent {
		if pid, ok := r.ClientParticipant[c.ID]; ok {
			if p, ok := r.Participants[pid]; ok {
				p.IsActive = false
			}
			delete(r.ClientParticipant, c.ID)
		}
		r.broadcastLocked(NewMessage(MsgPlayerLeft, map[string]string{
			"nickname": c.Nickname,
			"id":       c.ParticipantID.String(),
		}))
	}

	// If host disconnects, pause the game
	if c.Role == ClientHost && r.Status == models.RoomInProgress {
		r.Status = models.RoomPaused
		r.broadcastLocked(NewMessage(MsgGamePaused, map[string]string{
			"reason": "host_disconnected",
		}))
	}

	// Clean up empty room
	if len(r.Clients) == 0 {
		go r.Hub.DestroyRoom(r.PIN)
	}
}

// handleHostMessage dispatches host control messages
func (r *GameRoom) handleHostMessage(c *Client, msg Message) {
	switch msg.Type {
	case MsgStartGame:
		if r.Status != models.RoomWaiting {
			c.sendError("INVALID_STATE", "Game already started")
			return
		}
		r.Engine.Start(r)

	case MsgNextQuestion:
		if r.Status != models.RoomInProgress {
			c.sendError("INVALID_STATE", "Game is not in progress")
			return
		}
		r.Engine.NextQuestion(r)

	case MsgPauseGame:
		r.Engine.Pause(r)

	case MsgResumeGame:
		r.Engine.Resume(r)

	case MsgEndGame:
		r.Engine.End(r)

	case MsgKickPlayer:
		var p KickPlayerPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			c.sendError("INVALID_PAYLOAD", "Invalid payload")
			return
		}
		r.kickPlayer(p.ParticipantID)
	}
}

// handleStudentMessage dispatches student action messages
func (r *GameRoom) handleStudentMessage(c *Client, msg Message) {
	switch msg.Type {
	case MsgSubmitAnswer:
		var p SubmitAnswerPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			c.sendError("INVALID_PAYLOAD", "Invalid payload")
			return
		}
		r.mu.RLock()
		pid, ok := r.ClientParticipant[c.ID]
		r.mu.RUnlock()
		if !ok {
			return
		}
		r.Engine.HandleAnswer(r, pid, p)

	case MsgEmoji:
		var p EmojiPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			return
		}
		r.Broadcast(NewMessage(MsgEmojiReaction, map[string]string{
			"nickname": c.Nickname,
			"emoji":    p.Emoji,
		}))
	}
}

func (r *GameRoom) kickPlayer(participantIDStr string) {
	pid, err := uuid.Parse(participantIDStr)
	if err != nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for clientID, pID := range r.ClientParticipant {
		if pID == pid {
			if c, ok := r.Clients[clientID]; ok {
				c.sendError("KICKED", "You have been removed from the game")
				close(c.Send)
				delete(r.Clients, clientID)
			}
			delete(r.ClientParticipant, clientID)
			break
		}
	}
	delete(r.Participants, pid)
}

// Broadcast sends to all clients in the room
func (r *GameRoom) Broadcast(msg Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	r.broadcastLocked(msg)
}

// broadcastLocked — must hold r.mu.RLock
func (r *GameRoom) broadcastLocked(msg Message) {
	for _, c := range r.Clients {
		c.SendMessage(msg)
	}
}

// BroadcastStudents sends only to student clients
func (r *GameRoom) BroadcastStudents(msg Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.Clients {
		if c.Role == ClientStudent {
			c.SendMessage(msg)
		}
	}
}

// SendToParticipant sends to a specific participant
func (r *GameRoom) SendToParticipant(pid uuid.UUID, msg Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for clientID, pID := range r.ClientParticipant {
		if pID == pid {
			if c, ok := r.Clients[clientID]; ok {
				c.SendMessage(msg)
			}
			return
		}
	}
}

// SendToHost sends only to host client(s)
func (r *GameRoom) SendToHost(msg Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.Clients {
		if c.Role == ClientHost {
			c.SendMessage(msg)
		}
	}
}

func (r *GameRoom) buildRoomState() RoomStatePayload {
	players := make([]PlayerInfo, 0, len(r.Participants))
	for _, p := range r.Participants {
		players = append(players, playerToInfo(p))
	}

	quizTitle := ""
	hostName := ""
	totalQ := 0
	if r.Quiz != nil {
		quizTitle = r.Quiz.Title
		totalQ = len(r.Quiz.Questions)
	}

	return RoomStatePayload{
		RoomID:         r.RoomID.String(),
		PIN:            r.PIN,
		Status:         string(r.Status),
		GameMode:       string(r.GameMode),
		QuizTitle:      quizTitle,
		HostName:       hostName,
		Players:        players,
		TotalQuestions: totalQ,
	}
}

func playerToInfo(p *models.RoomParticipant) PlayerInfo {
	return PlayerInfo{
		ID:       p.ID.String(),
		Nickname: p.Nickname,
		Avatar:   p.Avatar,
		Score:    p.Score,
		Streak:   p.Streak,
		TeamID:   p.TeamID,
		IsActive: p.IsActive,
	}
}

func (r *GameRoom) GetParticipantCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Participants)
}

func (r *GameRoom) GetActiveParticipants() []*models.RoomParticipant {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.RoomParticipant, 0)
	for _, p := range r.Participants {
		if p.IsActive {
			result = append(result, p)
		}
	}
	return result
}

// AnsweredCount returns how many active students have answered the current question
func (r *GameRoom) AnsweredCount(questionID uuid.UUID, answered map[uuid.UUID]bool) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for pid, p := range r.Participants {
		if p.IsActive && answered[pid] {
			count++
		}
	}
	return count
}

func (r *GameRoom) Log(format string, args ...any) {
	log.Printf("[Room "+r.PIN+"] "+format, args...)
}
