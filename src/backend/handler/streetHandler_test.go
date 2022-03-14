package handler

import (
	"abfallkalender_api/src/backend/client"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type ClientMock struct {
	redirectURL  string
	houseNumbers []string
}

func (mt *ClientMock) GetRedirectUrl(_ string) (string, error) {
	return mt.redirectURL, nil
}

func (mt *ClientMock) GetHouseNumbers(_ string, _ string) (client.HouseNumbers, error) {
	numbers := client.HouseNumbers{}

	for _, number := range mt.houseNumbers {
		numbers = append(numbers, number)
	}

	return numbers, nil
}

func TestHappyPath(t *testing.T) {
	controller := Controller{
		Client: &ClientMock{
			redirectURL:  "www.mock.com/redirect",
			houseNumbers: []string{"2", "2-10"},
		},
	}

	streetName := "Aachener Straße"

	request := createTestRequest(streetName)
	writer := httptest.NewRecorder()

	controller.GetStreet(writer, request)

	res := writer.Result()
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	dto := streetWithHouseNumbersDto{}
	err = json.Unmarshal(data, &dto)

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
