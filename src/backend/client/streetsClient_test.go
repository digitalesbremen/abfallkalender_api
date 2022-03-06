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

	streets, _ := NewClient(server.BaseUrl).GetStreets(baseUrl)

	if len(streets) != 4 {
		t.Fatalf(`ReadStreets(%s) should contain %d entries but was %d`, server.BaseUrl, 4, len(streets))
	}
	if streets.notContains("Aachener Straße") {
		t.Fatalf(`ReadStreets(%s) should contain %s`, server.BaseUrl, "Aachener Straße")
	}
	if streets.notContains("Lars-Krüger-Hof") {
		t.Fatalf(`ReadStreets(%s) should contain %s`, server.BaseUrl, "Lars-Krüger-Hof")
	}
	if streets.notContains("Lars-Krüger-Hof") {
		t.Fatalf(`ReadStreets(%s) should contain %s`, server.BaseUrl, "Lars-Krüger-Hof")
	}
	if streets.notContains("Züricher Straße") {
		t.Fatalf(`ReadStreets(%s) should contain %s`, server.BaseUrl, "Züricher Straße")
	}
	if streets.contains("") {
		t.Fatalf(`ReadStreets(%s) should not contain empty string`, server.BaseUrl)
	}
}

func (r Streets) notContains(e string) bool {
	return !r.contains(e)
}

func (r Streets) contains(e string) bool {
	for _, a := range r {
		if a == e {
			return true
		}
	}
	return false
}
