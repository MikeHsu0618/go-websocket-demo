// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chatroom

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"otochat/internal/entity"
	"time"

	"github.com/gorilla/websocket"
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

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// allow origin
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub *Hub `json:"-"`

	// The websocket connection.
	Conn *websocket.Conn `json:"-"`

	// Buffered channel of outbound messages.
	Send chan entity.MessageResponse `json:"-"`

	Username string `json:"username"`
}

// ReadPump pumps messages from the websocket connection to the Hub.
//
// The application runs ReadPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		var msgReq entity.MessageRequest
		_ = json.Unmarshal(msg, &msgReq)
		if msgReq.Type == entity.OneToOne {
			c.Hub.oneToOne <- entity.NewMessageResponse(
				200,
				entity.OneToOne,
				msgReq.Msg,
				map[string]interface{}{
					"from_username": c.Username,
					"to_username":   msgReq.Data.(map[string]interface{})["to_username"].(string),
				})
		} else if msgReq.Type == entity.Broadcast {
			c.Hub.broadcast <- entity.NewMessageResponse(200, entity.Broadcast, msgReq.Msg, nil)
		}
	}
}

// WritePump pumps messages from the Hub to the websocket connection.
//
// A goroutine running WritePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
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
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.writeMessage(message)
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) writeMessage(message entity.MessageResponse) {
	msgJson, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return
	}
	w, err := c.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return
	}
	w.Write(msgJson)
	if message.Code == 403 {
		c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
		return
	}

	// Add queued chat messages to the current websocket message.
	n := len(c.Send)
	for i := 0; i < n; i++ {
		msgJson, err := json.Marshal(<-c.Send)
		if err != nil {
			log.Println(err)
			return
		}
		w.Write(newline)
		w.Write(msgJson)
		if message.Code == 403 {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
	}
	if err := w.Close(); err != nil {
		return
	}
}
