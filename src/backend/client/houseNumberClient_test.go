package client

import (
	"strings"
	"testing"
)

func TestGetHouseNumbers(t *testing.T) {
	server := startAbfallkalenderServer(t)

	defer server.Close()

	// TODO duplicate
	baseUrl := server.BaseUrl + strings.Replace(RedirectUrlHeader, "/Abfallkalender", "", 1)

	houseNumbers, _ := NewClient(server.BaseUrl).GetHouseNumbers(baseUrl, "Aachener+Stra%C3%9Fe")

	if len(houseNumbers) != 4 {
		t.Fatalf(`ReadStreets(%s) should contain %d entries but was %d`, server.BaseUrl, 4, len(houseNumbers))
	}
	if houseNumbers.notContains("0") {
		t.Fatalf(`GetHouseNumbers(%s) should contain %s`, server.BaseUrl, "0")
	}
	if houseNumbers.notContains("2") {
		t.Fatalf(`GetHouseNumbers(%s) should contain %s`, server.BaseUrl, "2")
	}
	if houseNumbers.notContains("2-10") {
		t.Fatalf(`GetHouseNumbers(%s) should contain %s`, server.BaseUrl, "2-10")
	}
	if houseNumbers.notContains("3") {
		t.Fatalf(`GetHouseNumbers(%s) should contain %s`, server.BaseUrl, "3")
	}
	if houseNumbers.contains("") {
		t.Fatalf(`GetHouseNumbers(%s) should not contain empty string`, server.BaseUrl)
	}
}

func (r HouseNumbers) notContains(e string) bool {
	return !r.contains(e)
}

func (r HouseNumbers) contains(e string) bool {
	for _, a := range r {
		if a == e {
			return true
		}
	}
	return false
}
