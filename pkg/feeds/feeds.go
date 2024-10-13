package feeds

import (
	"golang.org/x/text/language"
)

type FeedgenParams struct {
	Feed   string
	Limit  int64
	Offset string
	Langs  []language.Tag
}
