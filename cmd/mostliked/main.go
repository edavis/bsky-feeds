package main

import (
	"log"

	"github.com/edavis/bsky-feeds/pkg/mostliked"
	"github.com/gorilla/websocket"
)

func main() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:6008/subscribe?wantedCollections=app.bsky.feed.post&wantedCollections=app.bsky.feed.like", nil)
	if err != nil {
		log.Fatal("websocket connection error:", err)
	}
	defer conn.Close()

	jetstreamEvents := make(chan []byte)
	go mostliked.Handler(jetstreamEvents)

	log.Println("starting up")
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("error reading from websocket:", err)
			break
		}
		jetstreamEvents <- message
	}
}
