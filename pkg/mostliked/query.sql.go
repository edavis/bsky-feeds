// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package mostliked

import (
	"context"
)

const insertLang = `-- name: InsertLang :exec
insert or ignore into langs (uri, lang) values (?, ?)
`

type InsertLangParams struct {
	Uri  string
	Lang string
}

func (q *Queries) InsertLang(ctx context.Context, arg InsertLangParams) error {
	_, err := q.db.ExecContext(ctx, insertLang, arg.Uri, arg.Lang)
	return err
}

const insertPost = `-- name: InsertPost :exec
insert or ignore into posts (uri, create_ts, likes) values (?, ?, ?)
`

type InsertPostParams struct {
	Uri      string
	CreateTs int64
	Likes    int64
}

func (q *Queries) InsertPost(ctx context.Context, arg InsertPostParams) error {
	_, err := q.db.ExecContext(ctx, insertPost, arg.Uri, arg.CreateTs, arg.Likes)
	return err
}

const updateLikes = `-- name: UpdateLikes :exec
update posts set likes = likes + 1 where uri = ?
`

func (q *Queries) UpdateLikes(ctx context.Context, uri string) error {
	_, err := q.db.ExecContext(ctx, updateLikes, uri)
	return err
}
