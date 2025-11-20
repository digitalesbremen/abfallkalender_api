package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func (c Controller) GetStreets(w http.ResponseWriter, r *http.Request) {
	redirectUrl, err := c.Client.GetRedirectUrl(InitialContextPath)

	if err != nil {
		c.createInternalServerError(w, err)
		return
	}

	streets, err := c.Client.GetStreets(redirectUrl)

	if err != nil {
		c.createInternalServerError(w, err)
		return
	}

	var streetDtos []streetDto

	for _, streetName := range streets {
		streetDto := streetDto{}
		streetDto.Name = streetName
		streetDto.Links.Self.Href = buildStreetUrl(r, streetName)
		streetDtos = append(streetDtos, streetDto)
	}

	streetsDto := streetsDto{}
	streetsDto.Embedded.Streets = streetDtos

	dto, err := json.Marshal(streetsDto)

	if err != nil {
		c.createInternalServerError(w, err)
		return
	}

	// Set headers before writing status/body to ensure clients detect UTF-8 correctly
	w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(dto)
}

func buildStreetUrl(r *http.Request, streetName string) string {
	path := fmt.Sprintf("/abfallkalender-api/street/%s", url.QueryEscape(streetName))
	return absoluteURL(r, path)
}

type streetsDto struct {
	Embedded struct {
		Streets []streetDto `json:"streets"`
	} `json:"_embedded"`
}

type streetDto struct {
	Name  string `json:"name"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}
