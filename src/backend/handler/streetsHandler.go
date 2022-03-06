package handler

import (
	"abfallkalender_api/src/backend"
	"abfallkalender_api/src/backend/client"
	"encoding/json"
	"net/http"
)

func GetStreets(w http.ResponseWriter, _ *http.Request) {
	abfallkalenderClient := client.NewClient(api.BaseURL)
	// TODO handle error
	redirectUrl, _ := abfallkalenderClient.GetRedirectUrl(api.InitialContextPath)
	// TODO handle error
	streets, _ := abfallkalenderClient.GetStreets(redirectUrl)
	// TODO handle error
	dto, _ := json.Marshal(streets)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json; charset=utf-8")
	_, _ = w.Write(dto)
}
