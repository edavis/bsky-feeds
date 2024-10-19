package popular

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	appbsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/edavis/bsky-feeds/pkg/feeds"
)

type PostRow struct {
	Uri   string
	Score float64
}

func getPosts(ctx context.Context, dbCnx *sql.DB, langs []string, limit, offset int) ([]PostRow, error) {
	var queryParams []any
	var query strings.Builder
	fmt.Fprintf(&query, "select posts.uri, likes * exp( -1 * ( ( unixepoch('now') - create_ts ) / 1800.0 ) ) as score from posts")
	if len(langs) > 0 {
		fmt.Fprintf(&query, " left join langs on posts.uri = langs.uri where lang in (")
		for idx, lang := range langs {
			if idx > 0 {
				fmt.Fprintf(&query, ",")
			}
			fmt.Fprintf(&query, "?")
			queryParams = append(queryParams, lang)
		}
		fmt.Fprintf(&query, ")")
	}
	fmt.Fprintf(&query, " order by score desc limit ? offset ?")
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
		if err := rows.Scan(&post.Uri, &post.Score); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func Feed(params feeds.FeedgenParams) appbsky.FeedGetFeedSkeleton_Output {
	ctx := context.Background()
	dbCnx, err := sql.Open("sqlite3_custom", "data/mostliked.db?_journal=WAL&_fk=on")
	if err != nil {
		log.Printf("error opening db: %v\n", err)
	}
	defer dbCnx.Close()

	var langs []string
	for _, lang := range params.Langs {
		langs = append(langs, lang.String())
	}

	offset := 0
	if params.Cursor != "" {
		if parsed, err := strconv.Atoi(params.Cursor); err == nil {
			offset = parsed
		} else {
			log.Printf("error converting cursor: %v\n", err)
		}
	}

	rows, err := getPosts(ctx, dbCnx, langs, params.Limit, offset)
	if err != nil {
		log.Printf("error fetching rows: %v\n", err)
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
