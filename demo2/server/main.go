package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{} // use default options

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgrade:", err)
		return
	}
	defer conn.Close()

	// The event loop
	for {
		fmt.Println("來了喔")
		messageType, message, err := conn.ReadMessage()
		fmt.Println("沒塞")
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}
		log.Printf("Server Received: %s", message)
		err = conn.WriteMessage(messageType, []byte(`來自 Server 的關心`))
		if err != nil {
			log.Println("Error during message writing:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index Page")
}

func main() {
	http.HandleFunc("/socket", socketHandler)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
