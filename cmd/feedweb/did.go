package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const NgrokHostname = "routinely-right-barnacle.ngrok-free.app"

type DidDocument struct {
	context []string `json:"@context"`
	id string `json:"id"`
	services []DidService `json:"service"`
}

type DidService struct {
	id string `json:"id"`
	serviceType string `json:"type"`
	serviceEndpoint string `json:"serviceEndpoint"`
}

func didDoc(c echo.Context) error {
	doc := DidDocument{
		context: []string{"https://www.w3.org/ns/did/v1"},
		id: `did:web:` + NgrokHostname,
		services: []DidService{
			DidService{
				id: "#bsky_fg",
				serviceType: "BskyFeedGenerator",
				serviceEndpoint: `https://` + NgrokHostname,
			},
		},
	}
	return c.JSON(http.StatusOK, doc)
}
