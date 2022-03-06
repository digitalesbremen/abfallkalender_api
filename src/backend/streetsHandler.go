package api

import (
	"abfallkalender_api/src/backend/client"
	"encoding/json"
	"net/http"
)

func GetStreets(w http.ResponseWriter, _ *http.Request) {
	abfallkalenderClient := client.NewClient(BaseURL)
	// TODO handle error
	redirectUrl, _ := abfallkalenderClient.GetRedirectUrl(InitialContextPath)
	// TODO handle error
	streets, _ := abfallkalenderClient.GetStreets(redirectUrl)
	// TODO handle error
	dto, _ := json.Marshal(streets)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(dto)
}
