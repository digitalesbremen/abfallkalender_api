package handler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetStreetsHappyPath(t *testing.T) {
	data := sendGetStreetsRequest(t, controller)

	println(data)
}

func sendGetStreetsRequest(t *testing.T, controller Controller) []byte {
	request := createTestGetStreetsRequest()
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

func createTestGetStreetsRequest() *http.Request {
	testUrl := "http://www.mock.com/api/streets/"
	return httptest.NewRequest(http.MethodGet, testUrl, nil)
}
