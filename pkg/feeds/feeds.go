package feeds

import (
	"golang.org/x/text/language"
)

type FeedgenParams struct {
	Feed   string
	Limit  int
	Cursor string
	Langs  []language.Tag
}
