package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var clientMock = &ClientMock{
	redirectURL:  "www.mock.com/redirect",
	houseNumbers: []string{"2", "2-10"},
}

var controller = Controller{
	Client: clientMock,
}

func TestHappyPath(t *testing.T) {
	streetName := "Aachener Straße"

	data := sendRequest(t, controller, streetName)

	dto := streetWithHouseNumbersDto{}
	err := json.Unmarshal(data, &dto)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if dto.Name != streetName {
		t.Errorf("expected street %s got %s", streetName, dto.Name)
	}
	if dto.Links.Self.Href != "https://www.mock.com/api/street/Aachener+Stra%C3%9Fe" {
		t.Errorf("expected self link %s got %s", "https://www.mock.com/api/street/Aachener+Stra%C3%9Fe", dto.Links.Self.Href)
	}

	dto.verifyStreet(t, streetName)
	dto.verifyHouseNumber(t, streetName, "2")
	dto.verifyHouseNumber(t, streetName, "2-10")
}

func TestRedirectUrlReturnsError(t *testing.T) {
	clientMock.redirectError = errors.New("cannot get redirect URL")

	data := sendRequest(t, controller, "Aachener Straße")

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

func TestGetHouseNumbersReturnsError(t *testing.T) {
	clientMock.getHouseNumbersError = errors.New("cannot get house numbers")

	data := sendRequest(t, controller, "Aachener Straße")

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

func sendRequest(t *testing.T, controller Controller, streetName string) []byte {
	request := createTestRequest(streetName)
	writer := httptest.NewRecorder()

	controller.GetStreet(writer, request)

	res := writer.Result()
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	return data
}

func createTestRequest(streetName string) *http.Request {
	testUrl := "http://www.mock.com/api/streets/" + url.QueryEscape(streetName)
	request := httptest.NewRequest(http.MethodGet, testUrl, nil)

	// gorilla/mux add street name to vars
	vars := map[string]string{
		"street": streetName,
	}

	return mux.SetURLVars(request, vars)
}

func (dto streetWithHouseNumbersDto) verifyStreet(t *testing.T, street string) {
	if dto.Name != street {
		t.Errorf("expected street %s got %s", street, dto.Name)
	}

	expected := fmt.Sprintf("https://www.mock.com/api/street/%s", url.QueryEscape(street))

	if dto.Links.Self.Href != expected {
		t.Errorf("expected self link %s got %s", expected, dto.Links.Self.Href)
	}
}

func (dto streetWithHouseNumbersDto) verifyHouseNumber(t *testing.T, street, number string) {
	houseNumber := dto.getHouseNumber(number)

	if houseNumber == nil {
		t.Errorf(`house numbers should contain %s`, number)
	}

	expected := fmt.Sprintf("https://www.mock.com/api/street/%s/number/%s", url.QueryEscape(street), number)

	if houseNumber != nil && houseNumber.Links.Self.Href != expected {
		t.Errorf("expected house number self link %s got %s", expected, houseNumber.Links.Self.Href)
	}
}

func (dto streetWithHouseNumbersDto) getHouseNumber(number string) *houseNumberDto {
	for _, a := range dto.HouseNumbers {
		if a.Number == number {
			return &a
		}
	}

	return nil
}
