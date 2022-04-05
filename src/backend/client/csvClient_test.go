package client

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestGetCSV(t *testing.T) {
	server := startAbfallkalenderServer(t)

	defer server.Close()

	// TODO duplicate
	baseUrl := server.BaseUrl + strings.Replace(RedirectUrlHeader, "/Abfallkalender", "", 1)

	ical, _ := NewClient(server.BaseUrl).GetCSV(baseUrl, "Aachener+Stra%C3%9Fe", "22")

	response, _ := ioutil.ReadFile("test_csv_response.txt")

	if string(ical) != string(response) {
		t.Fatalf(`GetCSV(%s) should equal %s`, server.BaseUrl, string(response))
	}
}
