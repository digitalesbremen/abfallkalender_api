package client

import (
	"strings"
	"testing"
)

func TestGetStreets(t *testing.T) {
	server := startAbfallkalenderServer(t)

	defer server.Close()

	// TODO duplicate
	baseUrl := server.BaseUrl + strings.Replace(RedirectUrlHeader, "/Abfallkalender", "", 1)

	response, _ := NewClient(server.BaseUrl).GetStreets(baseUrl)

	if len(response.Streets) != 4 {
		t.Fatalf(`ReadStreets(%s) should contain %d entries but was %d`, server.BaseUrl, 4, len(response.Streets))
	}
	if response.notContains("Aachener Straße") {
		t.Fatalf(`ReadStreets(%s) should contain %s`, server.BaseUrl, "Aachener Straße")
	}
	if response.notContains("Lars-Krüger-Hof") {
		t.Fatalf(`ReadStreets(%s) should contain %s`, server.BaseUrl, "Lars-Krüger-Hof")
	}
	if response.notContains("Lars-Krüger-Hof") {
		t.Fatalf(`ReadStreets(%s) should contain %s`, server.BaseUrl, "Lars-Krüger-Hof")
	}
	if response.notContains("Züricher Straße") {
		t.Fatalf(`ReadStreets(%s) should contain %s`, server.BaseUrl, "Züricher Straße")
	}
	if response.contains("") {
		t.Fatalf(`ReadStreets(%s) should not contain empty string`, server.BaseUrl)
	}
}

func (r Response) notContains(e string) bool {
	return !r.contains(e)
}

func (r Response) contains(e string) bool {
	for _, a := range r.Streets {
		if a == e {
			return true
		}
	}
	return false
}
