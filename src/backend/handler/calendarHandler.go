package handler

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
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

	// TODO check different media types (ics, csv, pdf, html)

	ical, err := c.Client.GetICS(redirectUrl, url.QueryEscape(streetName), houseNumber)

	if err != nil {
		c.createInternalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "text/calendar; charset=utf-8")
	_, _ = w.Write(ical)
}

func parseHouseNumber(r *http.Request) string {
	params := mux.Vars(r)
	return params["number"]
}
