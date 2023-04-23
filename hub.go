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
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				client.send <- message
			}
			// TODO: Handle default case?
		}
	}
}
