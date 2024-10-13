package main

import (
	"golang.org/x/text/language"
	"net/http"

	appbsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/edavis/bsky-feeds/pkg/feeds"
	"github.com/edavis/bsky-feeds/pkg/mostliked"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type SkeletonRequest struct {
	Feed   string `query:"feed"`
	Limit  int  `query:"limit"`
	Cursor string `query:"cursor"`
}

type FeedLookup map[string]func(feeds.FeedgenParams) appbsky.FeedGetFeedSkeleton_Output

func parseLangs(userPrefs string) []language.Tag {
	t, _, _ := language.ParseAcceptLanguage(userPrefs)
	return t
}

func getFeedSkeleton(c echo.Context) error {
	var req SkeletonRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	generators := FeedLookup{
		"at://did:plc:4nsduwlpivpuur4mqkbfvm6a/app.bsky.feed.generator/most-liked": mostliked.Feed,
	}

	params := feeds.FeedgenParams{
		Feed:   req.Feed,
		Limit:  req.Limit,
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

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/.well-known/did.json", didDoc)
	e.GET("/xrpc/app.bsky.feed.getFeedSkeleton", getFeedSkeleton)
	e.Logger.Fatal(e.Start(":5000"))
}
