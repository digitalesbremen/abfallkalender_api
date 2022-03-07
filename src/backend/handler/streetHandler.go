package handler

import (
	"abfallkalender_api/src/backend/client"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type streetWithHouseNumbersDto struct {
	Name         string   `json:"name"`
	HouseNumbers []string `json:"houseNumbers"`
}

// GetStreet TODO test me
func GetStreet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	streetName := params["street"]

	abfallkalenderClient := client.NewClient(BaseURL)
	// TODO handle error
	redirectUrl, _ := abfallkalenderClient.GetRedirectUrl(InitialContextPath)
	// TODO handle error
	houseNumbers, _ := abfallkalenderClient.GetHouseNumbers(redirectUrl, streetName)

	var numbers []string

	for _, houseNumber := range houseNumbers {
		numbers = append(numbers, houseNumber)
	}

	streetsDto := streetWithHouseNumbersDto{
		Name:         streetName,
		HouseNumbers: numbers,
	}

	// TODO handle error
	dto, _ := json.Marshal(streetsDto)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json; charset=utf-8")
	_, _ = w.Write(dto)
}
