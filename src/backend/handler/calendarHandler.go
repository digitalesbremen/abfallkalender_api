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
	contentType := "text/calendar; charset=utf-8" // default to ICS

	// Check different media types via Accept header (ics, csv)
	switch acceptHeader := getAcceptHeader(r); acceptHeader {
	case NONE:
		response, err = c.Client.GetICS(redirectUrl, url.QueryEscape(streetName), houseNumber)
		contentType = "text/calendar; charset=utf-8"
	case ICS:
		response, err = c.Client.GetICS(redirectUrl, url.QueryEscape(streetName), houseNumber)
		contentType = "text/calendar; charset=utf-8"
	case CSV:
		response, err = c.Client.GetCSV(redirectUrl, url.QueryEscape(streetName), houseNumber)
		contentType = "text/csv; charset=utf-8"
	}

	if err != nil {
		c.createInternalServerError(w, err)
		return
	}

	// Set headers before writing status/body to ensure clients detect correctly
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

func parseHouseNumber(r *http.Request) string {
	params := mux.Vars(r)
	return params["number"]
}

func getAcceptHeader(r *http.Request) acceptHeader {
	if len(r.Header.Get("accept")) > 0 {
		for _, accept := range r.Header.Values("accept") {
			if strings.Contains(strings.ToLower(accept), "text/calendar") {
				return ICS
			}
			if strings.Contains(strings.ToLower(accept), "text/csv") {
				return CSV
			}
		}
	}

	return NONE
}

type acceptHeader int

const (
	NONE acceptHeader = iota
	ICS
	CSV
)
