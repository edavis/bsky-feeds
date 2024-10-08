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
	"github.com/karlseguin/ccache/v3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pemistahl/lingua-go"
)

//go:embed schema.sql
var ddl string

const MinLikes = 5

type DraftPost struct {
	Language lingua.Language
	Created  int64
	Likes    int64
}

func safeTimestamp(input string) int64 {
	utcNow := time.Now().UTC()
	if input == "" {
		return utcNow.Unix()
	}
	var t time.Time
	var err error
	layouts := []string{
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err = time.Parse(layout, input); err == nil {
			break
		}
	}
	if err != nil {
		return utcNow.Unix()
	}
	if t.Unix() <= 0 {
		return utcNow.Unix()
	} else if t.Add(-2*time.Minute).Compare(utcNow) == -1 {
		// accept as long as parsed time is no more than 2 minutes in the future
		return t.Unix()
	} else if t.Compare(utcNow) == 1 {
		return utcNow.Unix()
	} else {
		return utcNow.Unix()
	}
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

	if post.Text != "" {
		return post.Text
	} else if post.Embed.EmbedImages != nil {
		for _, image := range post.Embed.EmbedImages.Images {
			if image.Alt != "" {
				return image.Alt
			}
		}
	}
	return ""
}

func Handler(events <-chan []byte) {
	ctx := context.Background()

	dbCnx, err := sql.Open("sqlite3", "data/mostliked.db?_journal=WAL&_fk=on")
	if err != nil {
		log.Fatal("error opening db")
	}
	if _, err := dbCnx.ExecContext(ctx, ddl); err != nil {
		log.Fatal("couldn't create tables")
	}
	defer dbCnx.Close()

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

	for evt := range events {
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
				Created: safeTimestamp(post.CreatedAt),
				Likes:   0,
			}
			if text := findDetectableText(post); text != "" {
				language, _ := detector.DetectLanguageOf(text)
				draftPost.Language = language
			} else if len(post.Langs) > 0 {
				iso := lingua.GetIsoCode639_1FromValue(post.Langs[0])
				draftPost.Language = lingua.GetLanguageFromIsoCode639_1(iso)
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
			log.Println("storing", like.Subject.Uri, "in database")
			err := queries.InsertPost(ctx, db.InsertPostParams{
				Uri:      like.Subject.Uri,
				CreateTs: draftPost.Created,
				Likes:    draftPost.Likes,
			})
			if err != nil {
				log.Println("error inserting post")
			}
			err = queries.InsertLang(ctx, db.InsertLangParams{
				Uri:  like.Subject.Uri,
				Lang: strings.ToLower(draftPost.Language.IsoCode639_1().String()),
			})
			if err != nil {
				log.Println("error inserting lang")
			}
		} else {
			err := queries.UpdateLikes(ctx, like.Subject.Uri)
			if err != nil {
				log.Println("error updating likes")
			}
		}
	}
}