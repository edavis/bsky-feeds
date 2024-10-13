package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const NgrokHostname = "routinely-right-barnacle.ngrok-free.app"

type DidDocument struct {
	Context  []string     `json:"@context"`
	ID       string       `json:"id"`
	Services []DidService `json:"service"`
}

type DidService struct {
	ID              string `json:"id"`
	ServiceType     string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

func didDoc(c echo.Context) error {
	doc := DidDocument{
		Context: []string{"https://www.w3.org/ns/did/v1"},
		ID:      `did:web:` + NgrokHostname,
		Services: []DidService{
			DidService{
				ID:              "#bsky_fg",
				ServiceType:     "BskyFeedGenerator",
				ServiceEndpoint: `https://` + NgrokHostname,
			},
		},
	}
	return c.JSON(http.StatusOK, doc)
}
