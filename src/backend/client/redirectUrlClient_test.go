package client

import (
	"strings"
	"testing"
)

func TestGetRedirectUrl(t *testing.T) {
	server := startAbfallkalenderServer(t)

	defer server.Close()

	response, _ := NewClient(server.BaseUrl).GetRedirectUrl(RedirectUrlContextPath)

	if response != server.BaseUrl+strings.Replace(RedirectUrlHeader, "/Abfallkalender", "", 1) {
		t.Fatalf(`NewClient(%s).GetRedirectUrl(%s), got: %s, want: %s`, server.BaseUrl, RedirectUrlContextPath, response, RedirectUrlHeader)
	}
}
