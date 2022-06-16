package handler

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetCalendarHappyPathWithAcceptHeaderIsTextCalendar(t *testing.T) {
	controller.Client = &ClientMock{
		redirectURL: "www.mock.com/redirect",
		ics:         ([]byte)("some-ics-demo"),
	}

	streetName := "Aachener Straße"
	houseNumber := "22"

	data := sendGetCalendarRequest(t, controller, streetName, houseNumber, "text/calendar")

	if string(data) != "some-ics-demo" {
		t.Errorf("expected response to be %s got %s", "some-ics-demo", string(data))
	}
}

func TestGetCalendarHappyPathWithAcceptHeaderIsTextCsv(t *testing.T) {
	controller.Client = &ClientMock{
		redirectURL: "www.mock.com/redirect",
		csv:         ([]byte)("some-csv-demo"),
	}

	streetName := "Aachener Straße"
	houseNumber := "22"

	data := sendGetCalendarRequest(t, controller, streetName, houseNumber, "text/csv")

	if string(data) != "some-csv-demo" {
		t.Errorf("expected response to be %s got %s", "some-csv-demo", string(data))
	}
}

func TestGetCalendarHappyPathWithAcceptHeaderIsEmpty(t *testing.T) {
	controller.Client = &ClientMock{
		redirectURL: "www.mock.com/redirect",
		ics:         ([]byte)("some-ics-demo"),
	}

	streetName := "Aachener Straße"
	houseNumber := "22"

	data := sendGetCalendarRequest(t, controller, streetName, houseNumber, "")

	if string(data) != "some-ics-demo" {
		t.Errorf("expected response to be %s got %s", "some-ics-demo", string(data))
	}
}

func TestGetCalendarRedirectUrlReturnsError(t *testing.T) {
	controller.Client = &ClientMock{
		redirectError: errors.New("cannot get redirect URL"),
	}

	data := sendGetCalendarRequest(t, controller, "Aachener Straße", "22", "")

	dto := protocolError{}
	err := json.Unmarshal(data, &dto)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if dto.Code != 500 {
		t.Errorf("expected http code to be %d got %d", 500, dto.Code)
	}
	if dto.Message != "Internal Server Error" {
		t.Errorf("expected http error message to be %s got %s", "Internal Server Error", dto.Message)
	}
}

func TestGetCalendarGetICSReturnsError(t *testing.T) {
	controller.Client = &ClientMock{
		redirectURL: "www.mock.com/redirect",
		getICSError: errors.New("cannot get ICS"),
	}

	data := sendGetCalendarRequest(t, controller, "Aachener Straße", "22", "")

	dto := protocolError{}
	err := json.Unmarshal(data, &dto)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if dto.Code != 500 {
		t.Errorf("expected http code to be %d got %d", 500, dto.Code)
	}
	if dto.Message != "Internal Server Error" {
		t.Errorf("expected http error message to be %s got %s", "Internal Server Error", dto.Message)
	}
}

func sendGetCalendarRequest(t *testing.T, controller Controller, streetName string, houseNumber string, acceptHeader string) []byte {
	request := createTestGetCalendarRequest(streetName, houseNumber, acceptHeader)
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

func createTestGetCalendarRequest(streetName string, houseNumber string, acceptHeader string) *http.Request {
	testUrl := "http://www.mock.com/api/street/" + url.QueryEscape(streetName) + "/number/" + houseNumber
	request := httptest.NewRequest(http.MethodGet, testUrl, nil)
	if len(acceptHeader) > 0 {
		request.Header.Set("accept", acceptHeader)
	}

	// gorilla/mux add street name to vars
	vars := map[string]string{
		"street": streetName,
		"number": houseNumber,
	}

	return mux.SetURLVars(request, vars)
}
