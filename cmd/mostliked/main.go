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

const JetstreamUrl = `wss://jetstream2.us-west.bsky.network/subscribe?wantedCollections=app.bsky.feed.post&wantedCollections=app.bsky.feed.like`

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
		log.Printf("websocket closed\n")
	}()

	dbCnx, err := sql.Open("sqlite3", "data/mostliked.db?_journal=WAL&_fk=on&_timeout=5000&_sync=1&_txlock=immediate")
	if err != nil {
		log.Fatalf("failed to open database: %v\n", err)
	}
	defer func() {
		if _, err := dbCnx.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
			log.Printf("error doing final WAL checkpoint: %v\n", err)
		}
		if err := dbCnx.Close(); err != nil {
			log.Printf("failed to close db: %v\n", err)
		}
		log.Printf("db closed\n")
	}()

	jetstreamEvents := make(chan []byte)
	go mostliked.Handler(ctx, jetstreamEvents, dbCnx)

	log.Printf("starting up\n")
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				stop()
			}
			jetstreamEvents <- message
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down\n")
}
