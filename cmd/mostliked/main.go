package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/edavis/bsky-feeds/pkg/mostliked"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

// const JetstreamUrl = `wss://jetstream1.us-west.bsky.network/subscribe?wantedCollections=app.bsky.feed.post&wantedCollections=app.bsky.feed.like&cursor=1728846514000000`
const JetstreamUrl = `ws://localhost:6008/subscribe?wantedCollections=app.bsky.feed.post&wantedCollections=app.bsky.feed.like`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conn, _, err := websocket.DefaultDialer.Dial(JetstreamUrl, nil)
	if err != nil {
		log.Fatalf("failed to open websocket: %v\n", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close websocket: %v\n", err)
		}
		log.Println("websocket closed")
	}()

	dbCnx, err := sql.Open("sqlite3", "data/mostliked.db?_journal=WAL&_fk=on")
	if err != nil {
		log.Fatalf("failed to open database: %v\n", err)
	}
	defer func() {
		if err := dbCnx.Close(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
		log.Println("db closed")
	}()

	jetstreamEvents := make(chan []byte)
	go mostliked.Handler(ctx, jetstreamEvents, dbCnx)

	log.Println("starting up")
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			jetstreamEvents <- message
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")
}
