package mostliked

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	appbsky "github.com/bluesky-social/indigo/api/bsky"
	db "github.com/edavis/bsky-feeds/db/mostliked"
	"github.com/edavis/bsky-feeds/pkg/feeds"
	_ "github.com/mattn/go-sqlite3"
)

func Feed(params feeds.FeedgenParams) appbsky.FeedGetFeedSkeleton_Output {
	ctx := context.Background()
	dbCnx, err := sql.Open("sqlite3", "data/mostliked.db?_journal=WAL&_fk=on&mode=ro")
	if err != nil {
		log.Fatal("error opening db")
	}
	defer dbCnx.Close()

	offset, err := strconv.Atoi(params.Offset)
	if err != nil {
		log.Println("error converting offset to integer")
	}

	queries := db.New(dbCnx)
	rows, err := queries.ViewFeed(ctx, db.ViewFeedParams{
		Limit:  params.Limit,
		Offset: int64(offset),
	})
	if err != nil {
		log.Println("error fetching rows")
	}
	var cursor string
	posts := make([]*appbsky.FeedDefs_SkeletonFeedPost, 0, params.Limit)
	for _, row := range rows {
		posts = append(posts, &appbsky.FeedDefs_SkeletonFeedPost{Post: row.Uri})
		cursor = row.Uri
	}
	return appbsky.FeedGetFeedSkeleton_Output{
		Cursor: &cursor,
		Feed: posts,
	}
}
