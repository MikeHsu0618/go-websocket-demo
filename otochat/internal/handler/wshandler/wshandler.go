package wshandler

import (
	"log"
	"net/http"
	"net/url"
	"otochat/internal/entity"
	"otochat/internal/pkg/chatroom"
)

func NewHandler() {
	hub := chatroom.NewHub()
	go hub.Run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *chatroom.Hub, w http.ResponseWriter, r *http.Request) {
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Println(err)
		return
	}
	// validate username
	if query.Get("username") == "" {
		log.Println("Empty Username Error")
		return
	}

	conn, err := chatroom.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &chatroom.Client{
		Username: query.Get("username"),
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan entity.MessageResponse, 256),
	}
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}
