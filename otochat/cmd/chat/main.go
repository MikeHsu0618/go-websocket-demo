package main

import (
	"flag"
	"log"
	"net/http"
	"otochat/internal/handler/wshandler"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
	wshandler.NewHandler()
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
