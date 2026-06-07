package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

func parsePayload(msg Message, dst any) error {
	return json.Unmarshal(msg.Payload, dst)
}

// InboundMessage wraps a client message for the hub to process
type InboundMessage struct {
	Client  *Client
	Message Message
}

// JoinFunc is called when a student sends join_room message.
// Implemented by the RoomHandler which has DB access.
type JoinFunc func(c *Client, payload JoinRoomPayload) error

// Hub maintains all active WebSocket clients and game rooms
// Single goroutine owns all mutation — no locking needed on rooms/clients maps
type Hub struct {
	// All connected clients
	clients map[string]*Client

	// Active game rooms, keyed by PIN
	rooms map[string]*GameRoom

	// Channels
	Register   chan *Client
	Unregister chan *Client
	Inbound    chan InboundMessage

	// JoinHandler is set by RoomHandler to handle student join with DB access
	JoinHandler JoinFunc

	mu sync.RWMutex // only for external reads (e.g. stats endpoint)
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]*GameRoom),
		Register:   make(chan *Client, 64),
		Unregister: make(chan *Client, 64),
		Inbound:    make(chan InboundMessage, 512),
	}
}

// Run starts the hub event loop — call in a goroutine
func (h *Hub) Run() {
	log.Println("WebSocket Hub started")
	for {
		select {
		case client := <-h.Register:
			h.clients[client.ID] = client
			log.Printf("Client registered: %s (total: %d)", client.ID, len(h.clients))

		case client := <-h.Unregister:
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				h.handleClientDisconnect(client)
				log.Printf("Client unregistered: %s (total: %d)", client.ID, len(h.clients))
			}

		case msg := <-h.Inbound:
			h.handleMessage(msg.Client, msg.Message)
		}
	}
}

func (h *Hub) handleClientDisconnect(c *Client) {
	if c.Room == nil {
		return
	}
	c.Room.handleDisconnect(c)
}

func (h *Hub) handleJoinRoom(c *Client, msg Message) {
	var payload JoinRoomPayload
	if err := parsePayload(msg, &payload); err != nil {
		c.sendError("INVALID_PAYLOAD", "Invalid join payload")
		return
	}
	if h.JoinHandler == nil {
		c.sendError("SERVER_ERROR", "Join handler not configured")
		return
	}
	if err := h.JoinHandler(c, payload); err != nil {
		c.sendError("JOIN_FAILED", err.Error())
	}
}

func (h *Hub) handleMessage(c *Client, msg Message) {
	switch msg.Type {
	case MsgJoinRoom:
		h.handleJoinRoom(c, msg)

	case MsgStartGame, MsgNextQuestion, MsgPauseGame, MsgResumeGame, MsgEndGame, MsgKickPlayer:
		if c.Room == nil {
			c.sendError("NOT_IN_ROOM", "You are not in a room")
			return
		}
		if c.Role != ClientHost {
			c.sendError("NOT_HOST", "Only the host can do this")
			return
		}
		c.Room.handleHostMessage(c, msg)

	case MsgSubmitAnswer, MsgConfidencePick, MsgEmoji:
		if c.Room == nil {
			c.sendError("NOT_IN_ROOM", "You are not in a room")
			return
		}
		if c.Role != ClientStudent {
			return
		}
		c.Room.handleStudentMessage(c, msg)

	default:
		c.sendError("UNKNOWN_MESSAGE", "Unknown message type: "+msg.Type)
	}
}

// CreateRoom creates and registers a new game room
func (h *Hub) CreateRoom(room *GameRoom) {
	h.mu.Lock()
	h.rooms[room.PIN] = room
	h.mu.Unlock()
}

// GetRoom returns a room by PIN (thread-safe read)
func (h *Hub) GetRoom(pin string) (*GameRoom, bool) {
	h.mu.RLock()
	r, ok := h.rooms[pin]
	h.mu.RUnlock()
	return r, ok
}

// DestroyRoom removes a room from the hub
func (h *Hub) DestroyRoom(pin string) {
	h.mu.Lock()
	delete(h.rooms, pin)
	h.mu.Unlock()
	log.Printf("Room %s destroyed", pin)
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(room *GameRoom, msg Message) {
	room.mu.RLock()
	defer room.mu.RUnlock()
	for _, c := range room.Clients {
		c.SendMessage(msg)
	}
}

// SendToClient sends to a specific client
func (h *Hub) SendToClient(clientID string, msg Message) {
	if c, ok := h.clients[clientID]; ok {
		c.SendMessage(msg)
	}
}

func (h *Hub) Stats() map[string]int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return map[string]int{
		"clients": len(h.clients),
		"rooms":   len(h.rooms),
	}
}
