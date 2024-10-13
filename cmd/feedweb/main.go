package main

import (
	"golang.org/x/text/language"
	"net/http"
	"strconv"

	appbsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/edavis/bsky-feeds/pkg/feeds"
	"github.com/edavis/bsky-feeds/pkg/mostliked"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type SkeletonRequest struct {
	Feed   string `query:"feed"`
	Limit  string  `query:"limit"`
	Cursor string `query:"cursor"`
}

type FeedLookup map[string]func(feeds.FeedgenParams) appbsky.FeedGetFeedSkeleton_Output

func parseLangs(userPrefs string) []language.Tag {
	t, _, _ := language.ParseAcceptLanguage(userPrefs)
	return t
}

var generators = FeedLookup{
	"at://did:plc:4nsduwlpivpuur4mqkbfvm6a/app.bsky.feed.generator/most-liked": mostliked.Feed,
	"at://did:plc:4nsduwlpivpuur4mqkbfvm6a/app.bsky.feed.generator/most-liked-dev": mostliked.Feed,
}

func getFeedSkeleton(c echo.Context) error {
	var req SkeletonRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var limit int = 30
	if req.Limit != "" {
		if l, err := strconv.Atoi(req.Limit); err == nil {
			limit = l
		}
	}

	params := feeds.FeedgenParams{
		Feed:   req.Feed,
		Limit:  limit,
		Cursor: req.Cursor,
		Langs:  parseLangs(c.Request().Header.Get("Accept-Language")),
	}

	feedFunc, ok := generators[req.Feed]
	if !ok {
		return c.String(http.StatusNotFound, "feed not found")
	}
	feed := feedFunc(params)
	return c.JSON(http.StatusOK, feed)
}

func describeFeedGenerator(c echo.Context) error {
	type feed struct {
		URI syntax.ATURI `json:"uri"`
	}

	type gen struct {
		DID string `json:"did"`
		Feeds []feed `json:"feeds"`
	}

	out := gen{
		DID: `https://` + NgrokHostname,
	}

	for feedUri, _ := range generators {
		aturi, err := syntax.ParseATURI(feedUri)
		if err != nil {
			continue
		}
		out.Feeds = append(out.Feeds, feed{URI: aturi})
	}

	return c.JSON(http.StatusOK, out)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/.well-known/did.json", didDoc)
	e.GET("/xrpc/app.bsky.feed.getFeedSkeleton", getFeedSkeleton)
	e.GET("/xrpc/app.bsky.feed.describeFeedGenerator", describeFeedGenerator)
	e.Logger.Fatal(e.Start(":5000"))
}
