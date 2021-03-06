package client

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestGetICal(t *testing.T) {
	server := startAbfallkalenderServer(t)

	defer server.Close()

	// TODO duplicate
	baseUrl := server.BaseUrl + strings.Replace(RedirectUrlHeader, "/Abfallkalender", "", 1)

	ical, _ := NewClient(server.BaseUrl).GetICS(baseUrl, "Aachener+Stra%C3%9Fe", "22")

	response, _ := ioutil.ReadFile("test_ics_response.txt")

	if string(ical) != string(response) {
		t.Fatalf(`GetICS(%s) should equal %s`, server.BaseUrl, string(response))
	}
}
