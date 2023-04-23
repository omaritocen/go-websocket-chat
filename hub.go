package main

type Hub struct {
	// Registered Clients
	clients map[*Client]bool

	// Inbound messages from client
	broadcast chan []byte

	// Register/Unregister requests from client
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.clients[client] = true
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}
}

func (h *Hub) handleBroadcast(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:

		// If client buffer is full assume is dead or stuck
		default:
			h.unregisterClient(client)
		}
	}
}
