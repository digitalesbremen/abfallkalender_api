package handler

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetCalendarHappyPath(t *testing.T) {
	controller.Client = &ClientMock{
		redirectURL: "www.mock.com/redirect",
		ical:        "some-ical-demo",
	}

	streetName := "Aachener Straße"
	houseNumber := "22"

	data := sendGetCalendarRequest(t, controller, streetName, houseNumber)

	if string(data) != "some-ical-demo" {
		t.Errorf("expected response to be %s got %s", "some-ical-demo", string(data))
	}
}

func sendGetCalendarRequest(t *testing.T, controller Controller, streetName string, houseNumber string) []byte {
	request := createTestGetCalendarRequest(streetName, houseNumber)
	writer := httptest.NewRecorder()

	controller.GetCalendar(writer, request)

	res := writer.Result()
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	return data
}

func createTestGetCalendarRequest(streetName string, houseNumber string) *http.Request {
	testUrl := "http://www.mock.com/api/street/" + url.QueryEscape(streetName) + "/number/" + houseNumber
	request := httptest.NewRequest(http.MethodGet, testUrl, nil)

	// gorilla/mux add street name to vars
	vars := map[string]string{
		"street": streetName,
		"number": houseNumber,
	}

	return mux.SetURLVars(request, vars)
}
