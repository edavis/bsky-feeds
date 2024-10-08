package mostliked

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	db "github.com/edavis/bsky-feeds/db/mostliked"
	_ "github.com/mattn/go-sqlite3"
)

type FeedViewParams struct {
	Limit  int64
	Offset string
	Langs  []string
}

func Feed(args FeedViewParams) []string {
	ctx := context.Background()
	dbCnx, err := sql.Open("sqlite3", "data/mostliked.db?_journal=WAL&_fk=on")
	if err != nil {
		log.Fatal("error opening db")
	}
	defer dbCnx.Close()

	offset, err := strconv.Atoi(args.Offset)
	if err != nil {
		log.Println("error converting offset to integer")
	}

	queries := db.New(dbCnx)
	rows, err := queries.ViewFeed(ctx, db.ViewFeedParams{
		Limit:  args.Limit,
		Offset: int64(offset),
	})
	if err != nil {
		log.Println("error fetching rows")
	}
	var uris []string
	for _, row := range rows {
		uris = append(uris, row.Uri)
	}
	return uris
}
