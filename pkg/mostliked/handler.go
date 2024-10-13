package mostliked

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	appbsky "github.com/bluesky-social/indigo/api/bsky"
	jetstream "github.com/bluesky-social/jetstream/pkg/models"
	db "github.com/edavis/bsky-feeds/db/mostliked"
	"github.com/edavis/bsky-feeds/pkg/feeds"
	"github.com/karlseguin/ccache/v3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pemistahl/lingua-go"
)

//go:embed schema.sql
var ddl string

const MinLikes = 5

type DraftPost struct {
	Languages []lingua.Language
	Created   int64
	Likes     int64
}

func trimPostsTable(ctx context.Context, queries *db.Queries) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("trimming posts")
			if err := queries.TrimPosts(ctx); err != nil {
				log.Println("error trimming posts")
			}
		}
	}
}

func findDetectableText(post appbsky.FeedPost) string {
	// if we have text, detect against that
	// no text but we do have images, detect against first alt text
	// log.Printf("%+v\n", post)

	if post.Text != "" {
		return post.Text
	} else if post.Embed != nil && post.Embed.EmbedImages != nil && post.Embed.EmbedImages.Images != nil {
		for _, image := range post.Embed.EmbedImages.Images {
			if image.Alt != "" {
				return image.Alt
			}
		}
	}
	return ""
}

func Handler(ctx context.Context, events <-chan []byte, dbCnx *sql.DB) {
	if _, err := dbCnx.ExecContext(ctx, ddl); err != nil {
		log.Printf("couldn't create tables: %v\n", err)
	}
	queries := db.New(dbCnx)

	drafts := ccache.New(ccache.Configure[DraftPost]().MaxSize(50_000).GetsPerPromote(1))

	languages := []lingua.Language{
		lingua.Portuguese,
		lingua.English,
		lingua.Japanese,
		lingua.German,
		lingua.French,
		lingua.Spanish,
	}
	detector := lingua.
		NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		WithPreloadedLanguageModels().
		Build()

	go trimPostsTable(ctx, queries)

	var (
		dbTx       *sql.Tx
		queriesTx  *db.Queries
		txOpen     bool
		err        error
		eventCount int
	)

	for evt := range events {
		if !txOpen {
			dbTx, err = dbCnx.BeginTx(ctx, nil)
			if err != nil {
				log.Printf("failed to begin transaction: %v\n", err)
			}
			txOpen = true
			queriesTx = queries.WithTx(dbTx)
		}

		var like appbsky.FeedLike
		var event jetstream.Event
		if err := json.Unmarshal(evt, &event); err != nil {
			log.Println("error parsing jetstream event")
			continue
		}
		if event.Commit == nil {
			continue
		}
		commit := *event.Commit
		if commit.OpType != "c" {
			continue
		}
		if commit.Collection == "app.bsky.feed.post" {
			var post appbsky.FeedPost
			postUri := fmt.Sprintf("at://%s/%s/%s", event.Did, commit.Collection, commit.RKey)
			if err := json.Unmarshal(commit.Record, &post); err != nil {
				log.Println("error parsing appbsky.FeedPost")
				continue
			}
			draftPost := DraftPost{
				Created: feeds.SafeTimestamp(post.CreatedAt),
				Likes:   0,
			}
			if text := findDetectableText(post); text != "" {
				language, _ := detector.DetectLanguageOf(text)
				draftPost.Languages = []lingua.Language{language}
			} else if len(post.Langs) > 0 {
				var iso lingua.IsoCode639_1
				for _, lang := range post.Langs {
					iso = lingua.GetIsoCode639_1FromValue(lang)
					draftPost.Languages = append(draftPost.Languages, lingua.GetLanguageFromIsoCode639_1(iso))
				}
			}
			drafts.Set(postUri, draftPost, 30*time.Minute)
			continue
		} else if commit.Collection == "app.bsky.feed.like" {
			if err := json.Unmarshal(commit.Record, &like); err != nil {
				log.Println("error parsing appbsky.FeedLike")
				continue
			}
		}

		draft := drafts.Get(like.Subject.Uri)
		if draft != nil {
			draftPost := draft.Value()
			draftPost.Likes = draftPost.Likes + 1
			if draftPost.Likes < MinLikes {
				drafts.Replace(like.Subject.Uri, draftPost)
				continue
			}
			drafts.Delete(like.Subject.Uri)
			// log.Println("storing", like.Subject.Uri, "in database")
			err := queriesTx.InsertPost(ctx, db.InsertPostParams{
				Uri:      like.Subject.Uri,
				CreateTs: draftPost.Created,
				Likes:    draftPost.Likes,
			})
			if err != nil {
				log.Println("error inserting post")
			}
			for _, lang := range draftPost.Languages {
				err = queriesTx.InsertLang(ctx, db.InsertLangParams{
					Uri:  like.Subject.Uri,
					Lang: strings.ToLower(lang.IsoCode639_1().String()),
				})
				if err != nil {
					log.Println("error inserting lang")
				}
			}
		} else {
			err := queriesTx.UpdateLikes(ctx, like.Subject.Uri)
			if err != nil {
				log.Println("error updating likes")
			}
		}

		eventCount += 1
		if eventCount%1000 == 0 {
			if err := dbTx.Commit(); err != nil {
				log.Printf("commit failed: %v\n", err)
			} else {
				txOpen = false
			}
			// log.Println("db committed")
		}
	}
}
