package client

import (
	"strings"
	"testing"
)

func TestGetRedirectUrl(t *testing.T) {
	server := startAbfallkalenderServer(t)

	defer server.Close()

	got, _ := NewClient(server.BaseUrl).GetRedirectUrl(RedirectUrlContextPath)

	// TODO duplicate
	want := server.BaseUrl + strings.Replace(RedirectUrlHeader, "/Abfallkalender", "", 1)

	if got != want {
		t.Fatalf(`NewClient(%s).GetRedirectUrl(%s), got: %s, want: %s`, server.BaseUrl, RedirectUrlContextPath, got, want)
	}
}
