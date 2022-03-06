package handler

import (
	"abfallkalender_api/src/backend/client"
	"abfallkalender_api/src/backend/handler/model"
	"encoding/json"
	"net/http"
	"net/url"
)

func GetStreets(w http.ResponseWriter, r *http.Request) {
	abfallkalenderClient := client.NewClient(BaseURL)
	// TODO handle error
	redirectUrl, _ := abfallkalenderClient.GetRedirectUrl(InitialContextPath)
	// TODO handle error
	streets, _ := abfallkalenderClient.GetStreets(redirectUrl)

	println(r.Host)

	var streetDtos []model.StreetDto

	for _, streetName := range streets {
		streetDto := model.StreetDto{}
		streetDto.Name = streetName
		streetDto.Links.Self.Href = buildStreetUrl(r, streetName)
		streetDtos = append(streetDtos, streetDto)
	}

	streetsDto := model.StreetsDto{}
	streetsDto.Embedded.Streets = streetDtos

	// TODO handle error
	dto, _ := json.Marshal(streetsDto)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json; charset=utf-8")
	_, _ = w.Write(dto)
}

func buildStreetUrl(r *http.Request, streetName string) string {
	return "https://" + r.URL.Scheme + "://" + r.Host + "/api/street/" + url.QueryEscape(streetName)
}
