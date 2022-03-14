package handler

import (
	"abfallkalender_api/src/backend/client"
	"encoding/json"
	"fmt"
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

	redirectUrl, err := c.Client.GetRedirectUrl(InitialContextPath)

	if err != nil {
		c.createInternalServerError(w, err)
		return
	}

	houseNumbers, err := c.Client.GetHouseNumbers(redirectUrl, url.QueryEscape(streetName))

	if err != nil {
		c.createInternalServerError(w, err)
		return
	}

	if len(houseNumbers) == 0 {
		c.createNotFoundError(w, streetName, err)
		return
	}

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

	dto, err := json.Marshal(streetsDto)

	if err != nil {
		c.createInternalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json; charset=utf-8")
	_, _ = w.Write(dto)
}

func (c Controller) createInternalServerError(w http.ResponseWriter, err error) {
	fmt.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.
		NewEncoder(w).
		Encode(
			protocolError{
				Code:    http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			})
}

func (c Controller) createNotFoundError(w http.ResponseWriter, streetName string, err error) {
	fmt.Println(err)
	w.WriteHeader(http.StatusNotFound)
	_ = json.
		NewEncoder(w).
		Encode(
			protocolError{
				Code:    http.StatusNotFound,
				Message: fmt.Sprintf("Street '%s' or house numbers not found", streetName),
			})
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
