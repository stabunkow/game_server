package main

import "sync"

var defaultHub *Hub
var defaultHubOnce sync.Once

// Hub contains all gateway user clients, use uid as key
type Hub struct {
	clients    map[string]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func GetHub() *Hub {
	defaultHubOnce.Do(func() {
		defaultHub = newHub()
		go defaultHub.run()
	})
	return defaultHub
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.uid] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.uid]; ok {
				delete(h.clients, client.uid)
				close(client.send)
			}
		case message := <-h.broadcast:
			for uid, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, uid)
				}
			}
		}
	}
}

// Find hub client by uid
func (h *Hub) GetClient(uid string) *Client {
	client, ok := h.clients[uid]
	if !ok {
		return nil
	}
	return client
}
