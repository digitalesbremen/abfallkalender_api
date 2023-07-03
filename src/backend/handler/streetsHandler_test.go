package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetStreetsHappyPath(t *testing.T) {
	controller.Client = &ClientMock{
		redirectURL: "www.mock.com/redirect",
		streets:     []string{"Aachener Straße", "Eisenbahnerweg II (KG Grolland)"},
	}

	data := sendGetStreetsRequest(t, controller)

	dto := streetsDto{}
	err := json.Unmarshal(data, &dto)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if len(dto.Embedded.Streets) != 2 {
		t.Errorf("expected number of streets to be 2 got %v", len(dto.Embedded.Streets))
	}

	dto.verifyStreet(t, "Aachener Straße")
	dto.verifyStreet(t, "Eisenbahnerweg II (KG Grolland)")
}

func TestGetStreetsRedirectUrlClientReturnsError(t *testing.T) {
	controller.Client = &ClientMock{
		redirectError: errors.New("cannot get redirect URL"),
	}

	data := sendGetStreetsRequest(t, controller)

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

func TestGetStreetsClientReturnsError(t *testing.T) {
	controller.Client = &ClientMock{
		redirectURL:     "www.mock.com/redirect",
		getStreetsError: errors.New("cannot get streets"),
	}

	data := sendGetStreetsRequest(t, controller)

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

func sendGetStreetsRequest(t *testing.T, controller Controller) []byte {
	request := createTestGetStreetsRequest()
	writer := httptest.NewRecorder()

	controller.GetStreets(writer, request)

	res := writer.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	return data
}

func createTestGetStreetsRequest() *http.Request {
	testUrl := "http://www.mock.com/abfallkalender-api/streets/"
	return httptest.NewRequest(http.MethodGet, testUrl, nil)
}

func (dto streetsDto) verifyStreet(t *testing.T, streetName string) {
	street := dto.getStreet(streetName)

	if street == nil {
		t.Errorf(`streets should contain %s`, streetName)
	}

	expected := fmt.Sprintf("https://www.mock.com/abfallkalender-api/street/%s", url.QueryEscape(streetName))

	if street != nil && street.Links.Self.Href != expected {
		t.Errorf("expected street self link %s got %s", expected, street.Links.Self.Href)
	}
}

func (dto streetsDto) getStreet(streetName string) *streetDto {
	for _, street := range dto.Embedded.Streets {
		if street.Name == streetName {
			return &street
		}
	}

	return nil
}
