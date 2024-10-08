package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/edavis/bsky-feeds/pkg/mostliked"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:6008/subscribe?wantedCollections=app.bsky.feed.post&wantedCollections=app.bsky.feed.like", nil)
	if err != nil {
		log.Fatal("websocket connection error:", err)
	}
	defer conn.Close()

	dbCnx, err := sql.Open("sqlite3", "data/mostliked.db?_journal=WAL&_fk=on")
	if err != nil {
		log.Fatal("error opening db")
	}
	defer dbCnx.Close()

	jetstreamEvents := make(chan []byte)
	go mostliked.Handler(jetstreamEvents, dbCnx)

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func(conn *websocket.Conn, dbCnx *sql.DB, jetstreamEvents chan []byte) {
		<-signalChan
		log.Println("shutting down...")
		close(jetstreamEvents)
		dbCnx.Close()
		conn.Close()
		close(cleanupDone)
	}(conn, dbCnx, jetstreamEvents)

	log.Println("starting up")
	go func(conn *websocket.Conn) {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("error reading from websocket:", err)
				break
			}
			jetstreamEvents <- message
		}
	}(conn)

	<-cleanupDone
}
