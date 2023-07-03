package handler

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestGetNextHappyPath(t *testing.T) {
	csvFile, err := os.ReadFile("nextHandlerResponse.csv")

	if err != nil {
		t.Errorf("could not read '%v'", "nextHandlerResponse.csv")
	}

	controller.Client = &ClientMock{
		redirectURL: "www.mock.com/redirect",
		csv:         csvFile,
	}

	streetName := "Aachener Straße"
	houseNumber := "22"

	data := sendGetNextRequest(t, controller, streetName, houseNumber)

	responseWithoutNewLines := strings.ReplaceAll(string(data), "\n", "")
	expectedResponse := "{\"day_of_collection\":\"2045-01-02\",\"garbage_types\":[\"brown\",\"black\"]}"

	if responseWithoutNewLines != expectedResponse {
		t.Errorf("expected response to be %s got %s", expectedResponse, string(data))
	}
}

func TestGetNextRedirectUrlReturnsError(t *testing.T) {
	controller.Client = &ClientMock{
		redirectError: errors.New("cannot get redirect URL"),
	}

	data := sendGetNextRequest(t, controller, "Aachener Straße", "22")

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

func TestGetNextReturnsError(t *testing.T) {
	controller.Client = &ClientMock{
		redirectURL: "www.mock.com/redirect",
		getICSError: errors.New("cannot get CSV"),
	}

	data := sendGetNextRequest(t, controller, "Aachener Straße", "22")

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

func sendGetNextRequest(t *testing.T, controller Controller, streetName string, houseNumber string) []byte {
	request := createTestGetNextRequest(streetName, houseNumber)
	writer := httptest.NewRecorder()

	controller.GetNext(writer, request)

	res := writer.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	return data
}

func createTestGetNextRequest(streetName string, houseNumber string) *http.Request {
	testUrl := "http://www.mock.com/abfallkalender-api/street/" + url.QueryEscape(streetName) + "/number/" + houseNumber + "/next"
	request := httptest.NewRequest(http.MethodGet, testUrl, nil)

	// gorilla/mux add street name to vars
	vars := map[string]string{
		"street": streetName,
		"number": houseNumber,
	}

	return mux.SetURLVars(request, vars)
}
