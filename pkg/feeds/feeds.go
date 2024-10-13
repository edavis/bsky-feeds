package feeds

import (
	"golang.org/x/text/language"
)

type FeedgenParams struct {
	Feed   string
	Limit  int64
	Cursor string
	Langs  []language.Tag
}
