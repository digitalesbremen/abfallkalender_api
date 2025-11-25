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

// ClientCaller TODO move or split?
type ClientCaller interface {
	GetRedirectUrl(url string) (string, error)
	GetHouseNumbers(url string, streetName string) (client.HouseNumbers, error)
	GetStreets(redirectUrl string) (response client.Streets, err error)
	GetICS(redirectUrl string, streetName string, houseNumber string) (response []byte, err error)
	GetCSV(redirectUrl string, streetName string, houseNumber string) (response []byte, err error)
	GetLastCacheStatus() string
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
		// Base URL to the calendar resource (content is controlled via Accept header)
		baseCalURL := buildHouseNumberUrl(r, streetName, houseNumber)
		houseNumberDto.Links.Self.Href = baseCalURL
		// Keep payload lean: rely on content negotiation for ICS/CSV on self; expose only 'next' explicitly
		houseNumberDto.Links.Next.Href = baseCalURL + "/next"
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

	// Propagate cache status from last client call (if available)
	if cacheStatus := c.Client.GetLastCacheStatus(); cacheStatus != "" {
		w.Header().Set("X-Cache", cacheStatus)
	}
	// Set headers before writing status/body to ensure clients detect UTF-8 correctly
	w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(dto)
}

// TODO extract (used in multiple files
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
		// Computed "next" JSON endpoint; ICS/CSV are available via content negotiation on 'self'.
		Next struct {
			Href string `json:"href"`
			Type string `json:"type,omitempty"`
		} `json:"next"`
	} `json:"_links"`
}

func parseStreetName(r *http.Request) string {
	params := mux.Vars(r)
	return strings.Replace(params["street"], "+", " ", -1)
}

func buildHouseNumberUrl(r *http.Request, streetName string, houseNumber string) string {
	p := fmt.Sprintf("/abfallkalender-api/street/%s/number/%s", url.QueryEscape(streetName), url.QueryEscape(houseNumber))
	return absoluteURL(r, p)
}
