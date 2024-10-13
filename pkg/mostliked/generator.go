package mostliked

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"strings"
	"fmt"

	appbsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/edavis/bsky-feeds/pkg/feeds"
	_ "github.com/mattn/go-sqlite3"
)

type PostRow struct {
	Uri   string
}

func getPosts(ctx context.Context, dbCnx *sql.DB, langs []string, limit, offset int) ([]PostRow, error) {
	var queryParams []any
	var query strings.Builder
	fmt.Fprint(&query, "SELECT posts.uri FROM posts LEFT JOIN langs ON posts.uri = langs.uri")
	if len(langs) > 0 {
		fmt.Fprint(&query, " WHERE lang IN (")
		for idx, lang := range langs {
			if idx > 0 {
				fmt.Fprint(&query, ",")
			}
			fmt.Fprint(&query, "?")
			queryParams = append(queryParams, lang)
		}
		fmt.Fprint(&query, ")")
	}
	fmt.Fprint(&query, " ORDER BY likes DESC, create_ts DESC LIMIT ? OFFSET ?")
	queryParams = append(queryParams, limit, offset)

	fmt.Printf("query: %v %+v\n", query.String(), queryParams)

	rows, err := dbCnx.QueryContext(ctx, query.String(), queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []PostRow
	for rows.Next() {
		var post PostRow
		if err := rows.Scan(&post.Uri); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func Feed(params feeds.FeedgenParams) appbsky.FeedGetFeedSkeleton_Output {
	ctx := context.Background()
	dbCnx, err := sql.Open("sqlite3", "data/mostliked.db?_journal=WAL&_fk=on&mode=ro")
	if err != nil {
		log.Printf("error opening db: %v\n", err)
	}
	defer dbCnx.Close()

	var langs []string
	for _, lang := range params.Langs {
		langs = append(langs, lang.String())
	}

	offset := 0
	if parsed, err := strconv.Atoi(params.Cursor); err == nil {
		offset = parsed
	} else {
		log.Println("error converting cursor")
	}

	rows, err := getPosts(ctx, dbCnx, langs, params.Limit, offset)
	if err != nil {
		log.Println("error fetching rows:", err)
	}

	var cursor string
	posts := make([]*appbsky.FeedDefs_SkeletonFeedPost, 0, params.Limit)

	for _, row := range rows {
		posts = append(posts, &appbsky.FeedDefs_SkeletonFeedPost{Post: row.Uri})
	}

	skeleton := appbsky.FeedGetFeedSkeleton_Output{
		Feed: posts,
	}

	offset += len(posts)
	cursor = strconv.Itoa(offset)

	if len(posts) == params.Limit {
		skeleton.Cursor = &cursor
	}

	return skeleton
}
