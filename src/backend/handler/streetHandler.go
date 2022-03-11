package handler

import (
	"abfallkalender_api/src/backend/client"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strings"
)

type ClientCaller interface {
	GetRedirectUrl(url string) (string, error)
	GetHouseNumbers(url string, streetName string) (client.HouseNumbers, error)
}

type Controller struct {
	Client ClientCaller
}

func NewController() Controller {
	return Controller{
		Client: client.NewClient(BaseURL),
	}
}

func (c Controller) GetStreet(w http.ResponseWriter, r *http.Request) {
	streetName := parseStreetName(r)

	// TODO handle error
	redirectUrl, _ := c.Client.GetRedirectUrl(InitialContextPath)
	// TODO handle error
	// TODO handle houseNumbers are empty -> 404?
	houseNumbers, _ := c.Client.GetHouseNumbers(redirectUrl, url.QueryEscape(streetName))

	var numbers []houseNumberDto

	for _, houseNumber := range houseNumbers {
		houseNumberDto := houseNumberDto{}
		houseNumberDto.Number = houseNumber
		houseNumberDto.Links.Self.Href = buildHouseNumberUrl(r, streetName, houseNumber)
		numbers = append(numbers, houseNumberDto)
	}

	streetsDto := streetWithHouseNumbersDto{
		Name:         streetName,
		HouseNumbers: numbers,
	}

	streetsDto.Links.Self.Href = buildStreetUrl(r, streetName)

	// TODO handle error
	dto, _ := json.Marshal(streetsDto)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json; charset=utf-8")
	_, _ = w.Write(dto)
}

type streetWithHouseNumbersDto struct {
	Name         string           `json:"name"`
	HouseNumbers []houseNumberDto `json:"houseNumbers"`
	Links        struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

type houseNumberDto struct {
	Number string `json:"number"`
	Links  struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

func parseStreetName(r *http.Request) string {
	params := mux.Vars(r)
	return strings.Replace(params["street"], "+", " ", -1)
}

func buildHouseNumberUrl(r *http.Request, streetName string, houseNumber string) string {
	// TODO use fmt.printf
	return "https://" + r.Host + "/api/street/" + url.QueryEscape(streetName) + "/number/" + url.QueryEscape(houseNumber)
}
