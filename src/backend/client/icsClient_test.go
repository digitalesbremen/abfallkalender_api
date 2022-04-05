package client

import (
	"io/ioutil"
	"testing"
)

func TestGetICal(t *testing.T) {
	server := startAbfallkalenderServer(t)

	defer server.Close()

	// TODO duplicate
	baseUrl := server.BaseUrl + RedirectUrlHeader

	ical, _ := NewClient(server.BaseUrl).GetICS(baseUrl, "Aachener+Stra%C3%9Fe", "22")

	icalResponse, _ := ioutil.ReadFile("test_ical_response.txt")

	if ical != string(icalResponse) {
		t.Fatalf(`GetICS(%s) should equal %s`, server.BaseUrl, string(icalResponse))
	}
}
