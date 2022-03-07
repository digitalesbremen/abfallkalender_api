package handler

import (
	"abfallkalender_api/src/backend/client"
	"encoding/json"
	"net/http"
	"net/url"
)

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

// GetStreets TODO test me
func GetStreets(w http.ResponseWriter, r *http.Request) {
	abfallkalenderClient := client.NewClient(BaseURL)
	// TODO handle error
	redirectUrl, _ := abfallkalenderClient.GetRedirectUrl(InitialContextPath)
	// TODO handle error
	streets, _ := abfallkalenderClient.GetStreets(redirectUrl)

	var streetDtos []streetDto

	for _, streetName := range streets {
		streetDto := streetDto{}
		streetDto.Name = streetName
		streetDto.Links.Self.Href = buildStreetUrl(r, streetName)
		streetDtos = append(streetDtos, streetDto)
	}

	streetsDto := streetsDto{}
	streetsDto.Embedded.Streets = streetDtos

	// TODO handle error
	dto, _ := json.Marshal(streetsDto)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json; charset=utf-8")
	_, _ = w.Write(dto)
}

func buildStreetUrl(r *http.Request, streetName string) string {
	// TODO use fmt.printf
	return "https://" + r.Host + "/api/street/" + url.QueryEscape(streetName)
}
