package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

	// TODO verify dto
}

func sendGetStreetsRequest(t *testing.T, controller Controller) []byte {
	request := createTestGetStreetsRequest()
	writer := httptest.NewRecorder()

	controller.GetStreets(writer, request)

	res := writer.Result()
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	return data
}

func createTestGetStreetsRequest() *http.Request {
	testUrl := "http://www.mock.com/api/streets/"
	return httptest.NewRequest(http.MethodGet, testUrl, nil)
}
