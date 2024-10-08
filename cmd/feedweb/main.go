package main

import (
	"net/http"

	appbsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/edavis/bsky-feeds/pkg/mostliked"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type SkeletonRequest struct {
	Feed   string `query:"feed"`
	Limit  int64  `query:"limit"`
	Offset string `query:"offset"`
}

type SkeletonHeader struct {
	Langs []string `header:"Accept-Language"`
}

func getFeedSkeleton(c echo.Context) error {
	var req SkeletonRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	var hdr SkeletonHeader
	if err := (&echo.DefaultBinder{}).BindHeaders(c, &hdr); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var posts []*appbsky.FeedDefs_SkeletonFeedPost
	uris := mostliked.Feed(mostliked.FeedViewParams{
		Limit:  req.Limit,
		Offset: req.Offset,
		Langs:  hdr.Langs,
	})
	for _, uri := range uris {
		posts = append(posts, &appbsky.FeedDefs_SkeletonFeedPost{Post: uri})
	}

	return c.JSON(http.StatusOK, appbsky.FeedGetFeedSkeleton_Output{
		Feed: posts,
	})
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/xrpc/app.bsky.feed.getFeedSkeleton", getFeedSkeleton)
	e.Logger.Fatal(e.Start(":5000"))
}
