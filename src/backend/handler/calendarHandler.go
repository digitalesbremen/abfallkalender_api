package handler

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strings"
)

func (c Controller) GetCalendar(w http.ResponseWriter, r *http.Request) {
	streetName := parseStreetName(r)
	houseNumber := parseHouseNumber(r)

	redirectUrl, err := c.Client.GetRedirectUrl(InitialContextPath)

	if err != nil {
		// TODO handle 404
		c.createInternalServerError(w, err)
		return
	}

	var response []byte

	// TODO check different media types (ics, csv, pdf, html)
	switch acceptHeader := getAcceptHeader(r); acceptHeader {
	case NONE:
		response, err = c.Client.GetICS(redirectUrl, url.QueryEscape(streetName), houseNumber)
	case ICS:
		response, err = c.Client.GetICS(redirectUrl, url.QueryEscape(streetName), houseNumber)
	}

	if err != nil {
		c.createInternalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "text/calendar; charset=utf-8")
	_, _ = w.Write(response)
}

func parseHouseNumber(r *http.Request) string {
	params := mux.Vars(r)
	return params["number"]
}

func getAcceptHeader(r *http.Request) acceptHeader {
	if len(r.Header.Get("accept")) > 0 {
		for _, accept := range r.Header.Values("accept") {
			if strings.Contains(strings.ToLower(accept), strings.ToLower("text/calendar")) {
				return ICS
			}
		}
	}

	return NONE
}

type acceptHeader int

const (
	NONE acceptHeader = iota
	ICS
)
