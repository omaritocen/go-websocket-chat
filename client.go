package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub *Hub

	// Websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Name of the client
	name string

	// unique id for the client
	id string
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.handleNewMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func newClient(hub *Hub, conn *websocket.Conn, name string) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		name: name,
		id:   uuid.NewString(),
	}
}

func (c *Client) handleNewMessage(jsonMessage []byte) {
	message := decodeMessage(jsonMessage)
	message.Author = c

	switch message.Action {
	case JoinRoomAction:
		c.handleJoinRoomMessage(message)
	case LeaveRoomAction:
		c.handleLeaveRoomMessage(message)
	case TextMessageAction:
		c.handleTextMessage(message)
	}
}

func (c *Client) handleJoinRoomMessage(message Message) {
	roomName := message.Body
	room := c.hub.getRoomByName(roomName)

	if room == nil {
		room = c.hub.createRoom(roomName)
		go room.Run()
		log.Printf("Created new room, Name: [%s], id: [%s]\n", room.name, room.id)
	}

	room.register <- c
}

func (c *Client) handleLeaveRoomMessage(message Message) {
	roomName := message.Body
	room := c.hub.getRoomByName(roomName)

	if room != nil {
		room.unregister <- c
	}
}

func (c *Client) handleTextMessage(message Message) {
	roomId := message.Target
	room := c.hub.getRoomById(roomId)

	if room == nil {
		return
	}

	room.broadcast <- &message
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	clientName := r.URL.Query().Get("name")
	if clientName == "" {
		clientName = "NewUser"
	}

	client := newClient(hub, conn, clientName)
	log.Printf("New user joined the system, Name: [%s], ID: [%s]\n", client.name, client.id)

	hub.register <- client

	go client.writePump()
	go client.readPump()
}
