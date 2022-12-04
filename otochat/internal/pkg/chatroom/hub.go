// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chatroom

import (
	"otochat/internal/entity"
	"strings"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	broadcast chan entity.MessageResponse

	// Inbound messages for one on one.
	oneToOne chan entity.MessageResponse

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// mutex
	mux sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		broadcast:  make(chan entity.MessageResponse),
		oneToOne:   make(chan entity.MessageResponse),
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mux.Lock()
	defer h.mux.Unlock()
	if _, ok := h.clients[client.Username]; ok || client.Username == "" {
		client.Send <- entity.NewMessageResponse(403, entity.OneToOne, "Existed Username: "+client.Username, nil)
		return
	}
	h.clients[client.Username] = client
	h.notifyClients()
}

func (h *Hub) unregisterClient(client *Client) {
	h.mux.Lock()
	defer h.mux.Unlock()
	if c, ok := h.clients[client.Username]; ok {
		if client == c {
			delete(h.clients, client.Username)
		}
		close(client.Send)
	}
	h.notifyClients()
}

func (h *Hub) notifyClients() {
	// get user list while anyone online or offline
	users := make([]*Client, 0)
	userStrArr := make([]string, 0)
	for username, _ := range h.clients {
		userStrArr = append(userStrArr, username)
		users = append(users, &Client{Username: username})
	}
	msg := entity.NewMessageResponse(
		200,
		entity.OnlineUsers,
		"當前上線使用者: "+strings.Join(userStrArr, ", "),
		map[string]interface{}{"users": users})
	for _, client := range h.clients {
		client.Send <- msg
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			for username, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, username)
				}
			}
		case message := <-h.oneToOne:
			// only sent to each other
			for username, client := range h.clients {
				if username != message.Data.(map[string]interface{})["to_username"].(string) &&
					username != message.Data.(map[string]interface{})["from_username"].(string) {
					continue
				}
				client.Send <- message
			}
		}
	}
}
