package client

import (
	"testing"
)

func TestGetICal(t *testing.T) {
	server := startAbfallkalenderServer(t)

	defer server.Close()

	// TODO duplicate
	baseUrl := server.BaseUrl + RedirectUrlHeader

	ical, _ := NewClient(server.BaseUrl).GetICal(baseUrl, "Aachener+Stra%C3%9Fe", "22")

	if ical != icalResponse {
		t.Fatalf(`GetICal(%s) should equal %s`, server.BaseUrl, icalResponse)
	}
}
