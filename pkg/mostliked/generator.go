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
	Likes int
}

func getPosts(ctx context.Context, dbCnx *sql.DB, params feeds.FeedgenParams) ([]PostRow, error) {
	var queryParams []any
	var query strings.Builder
	fmt.Fprint(&query, "SELECT posts.uri, likes FROM posts LEFT JOIN langs ON posts.uri = langs.uri")
	fmt.Fprint(&query, " WHERE 1=1 ")
	if len(params.Langs) > 0 {
		fmt.Fprint(&query, " AND ( ")
		for idx, lang := range params.Langs {
			if idx > 0 {
				fmt.Fprint(&query, " OR ")
			}
			fmt.Fprint(&query, " lang = ? ")
			queryParams = append(queryParams, lang.String())
		}
		fmt.Fprint(&query, " ) ")
	}
	if params.Cursor != "" {
		fmt.Fprint(&query, " AND likes <= ? ")
		queryParams = append(queryParams, params.Cursor)
	}
	fmt.Fprint(&query, "ORDER BY likes DESC, create_ts DESC")
	fmt.Fprint(&query, "LIMIT ?")
	queryParams = append(queryParams, params.Limit+1)

	rows, err := dbCnx.QueryContext(ctx, query.String(), queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []PostRow
	for rows.Next() {
		var post PostRow
		if err := rows.Scan(&post.Uri, &post.Likes); err != nil {
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
		log.Fatal("error opening db")
	}
	defer dbCnx.Close()

	rows, err := getPosts(ctx, dbCnx, params)
	if err != nil {
		log.Println("error fetching rows:", err)
	}

	var cursor string
	posts := make([]*appbsky.FeedDefs_SkeletonFeedPost, 0, params.Limit)

	for idx, row := range rows {
		if idx+1 <= params.Limit {
			posts = append(posts, &appbsky.FeedDefs_SkeletonFeedPost{Post: row.Uri})
		}
		cursor = strconv.Itoa(row.Likes)
	}

	skeleton := appbsky.FeedGetFeedSkeleton_Output{
		Feed: posts,
	}

	if len(rows) == params.Limit {
		skeleton.Cursor = &cursor
	}

	return skeleton
}
