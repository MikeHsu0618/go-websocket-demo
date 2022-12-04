// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	broadcast chan MessageResponse

	// Inbound messages for one on one.
	oneToOne chan MessageResponse

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// mutex
	mux sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		broadcast:  make(chan MessageResponse),
		oneToOne:   make(chan MessageResponse),
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mux.Lock()
	defer h.mux.Unlock()
	if _, ok := h.clients[client.Username]; ok || client.Username == "" {
		msg := MessageResponse{
			Code: 404,
			Type: OneToOne,
			Msg:  "Existed Username: " + client.Username,
			Data: nil,
		}
		client.send <- msg
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
		close(client.send)
	}
	h.notifyClients()
}

func (h *Hub) notifyClients() {
	// 登入時取得上線人數
	fmt.Println(len(h.clients))
	users := make([]Client, 0)
	userStrArr := make([]string, 0)
	for username, _ := range h.clients {
		userStrArr = append(userStrArr, username)
		users = append(users, Client{Username: username})
	}
	msg := MessageResponse{
		Code: 200,
		Type: OnlineUsers,
		Msg:  "當前上線使用者: " + strings.Join(userStrArr, ", "),
		Data: map[string]interface{}{
			"users": users,
		},
	}
	for _, client := range h.clients {
		client.send <- msg
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			for username, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, username)
				}
			}
		case message := <-h.oneToOne:
			// 只傳訊息給 one to one 雙方
			for username, client := range h.clients {
				if username != message.Data.(map[string]interface{})["to_username"].(string) &&
					username != message.Data.(map[string]interface{})["from_username"].(string) {
					continue
				}
				client.send <- message
			}
		}
	}
}
