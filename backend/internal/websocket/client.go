package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type ClientRole string

const (
	ClientHost    ClientRole = "host"
	ClientStudent ClientRole = "student"
)

// Client represents a single WebSocket connection
type Client struct {
	ID            string
	Conn          *websocket.Conn
	Hub           *Hub
	Room          *GameRoom
	Role          ClientRole
	ParticipantID uuid.UUID
	UserID        *uuid.UUID
	Nickname      string
	Avatar        string
	Send          chan []byte
	Done          chan struct{}
}

func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:   uuid.New().String(),
		Conn: conn,
		Hub:  hub,
		Send: make(chan []byte, 256),
		Done: make(chan struct{}),
	}
}

// ReadPump pumps messages from WebSocket to hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
		close(c.Done)
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("WebSocket error for client %s: %v", c.ID, err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			c.sendError("INVALID_MESSAGE", "Invalid message format")
			continue
		}

		c.Hub.Inbound <- InboundMessage{Client: c, Message: msg}
	}
}

// WritePump pumps messages from hub to WebSocket
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.Done:
			return
		}
	}
}

func (c *Client) sendError(code, message string) {
	msg := NewMessage(MsgError, ErrorPayload{Code: code, Message: message})
	data, _ := json.Marshal(msg)
	select {
	case c.Send <- data:
	default:
	}
}

func (c *Client) SendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.Send <- data:
	default:
		log.Printf("Client %s send buffer full, dropping message", c.ID)
	}
}
