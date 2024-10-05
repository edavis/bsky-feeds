-- name: InsertPost :exec
insert or ignore into posts (uri, create_ts, likes) values (?, ?, ?);

-- name: InsertLang :exec
insert or ignore into langs (uri, lang) values (?, ?);

-- name: UpdateLikes :exec
update posts set likes = likes + 1 where uri = ?;

-- name: TrimPosts :exec
delete from posts where create_ts < unixepoch('now', '-24 hours');
