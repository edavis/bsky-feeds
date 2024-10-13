package mostliked

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	appbsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/edavis/bsky-feeds/pkg/feeds"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sanggonlee/gosq"
)

type PostRow struct {
	Uri   string
	Likes int
}

func getPosts(ctx context.Context, dbCnx *sql.DB, params feeds.FeedgenParams) ([]PostRow, error) {
	q, err := gosq.Execute(`
	SELECT posts.uri, likes
	FROM posts LEFT JOIN langs ON posts.uri = langs.uri
	WHERE 1=1
	{{if .Langs}}
		AND (
			{{- range $index, $lang := .Langs }}
			{{- if $index }} OR {{ end }} lang = ?
			{{- end }}
		)
	{{end}}
	{{if .Offset}}
	AND likes < ?
	{{end}}
	ORDER BY likes DESC, create_ts DESC
	LIMIT ?
	`, params)
	if err != nil {
		log.Println("error in sql template:", err)
	}

	log.Println(q)

	var queryParams []any
	for _, lang := range params.Langs {
		queryParams = append(queryParams, lang.String())
	}

	if params.Offset != "" {
		queryParams = append(queryParams, params.Offset)
	}

	queryParams = append(queryParams, params.Limit)

	log.Println(queryParams)

	rows, err := dbCnx.QueryContext(ctx, q, queryParams...)
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

	for _, row := range rows {
		posts = append(posts, &appbsky.FeedDefs_SkeletonFeedPost{Post: row.Uri})
		cursor = strconv.Itoa(row.Likes)
	}

	skeleton := appbsky.FeedGetFeedSkeleton_Output{
		Feed: posts,
	}

	if len(rows) == int(params.Limit) {
		skeleton.Cursor = &cursor
	}

	return skeleton
}
