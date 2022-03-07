package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	RedirectUrlContextPath  = "/bremenabfallkalender/Abfallkalender"
	RedirectUrlResponse     = "<html><head><title>Object moved</title></head><body>\n<h2>Object moved to <a href=\"/bremenabfallkalender/(S(nni))/Abfallkalender\">here</a>.</h2>\n</body></html>"
	RedirectUrlHeader       = "/bremenabfallkalender/(S(nni))/Abfallkalender"
	streetsContextPath      = "/bremenabfallkalender/(S(nni))/Data/Strassen"
	houseNumbersContextPath = "/bremenabfallkalender/(S(nni))/Data/Hausnummern?strasse=Aachener+Stra%C3%9Fe"
	streetsResponse         = "[\"\",\n\"Aachener Straße\",\"Lars-Krüger-Hof\",\"Martinsweg (KG Gartenstadt Vahr)\",\n\"Züricher Straße\"]"
	houseNumbersResponse    = "[\"\",\n\"0\",\"2\",\"2-10\",\n\"3\"]"
)

type AbfallkalenderServer struct {
	server             *httptest.Server
	BaseUrl            string
	StreetsContextPath string
}

func (s *AbfallkalenderServer) Close() {
	s.server.Close()
}

func startAbfallkalenderServer(t *testing.T) AbfallkalenderServer {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.String() {
		case streetsContextPath:
			doGetStreets(t, rw, req)
			break
		case houseNumbersContextPath:
			doGetHouseNumbers(t, rw, req)
			break
		case RedirectUrlContextPath:
			doGetServerRedirectUrl(t, rw, req)
			break
		default:
			_ = fmt.Sprintf("URL %s not known on test server", req.URL.String())
			t.FailNow()
		}
	}))

	return AbfallkalenderServer{server: server, BaseUrl: server.URL, StreetsContextPath: streetsContextPath}
}

func doGetStreets(t *testing.T, rw http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		_ = fmt.Sprintf("%s %s, want: GET", req.Method, streetsContextPath)
		t.FailNow()
	}

	_, _ = rw.Write([]byte(streetsResponse))
}

func doGetHouseNumbers(t *testing.T, rw http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		_ = fmt.Sprintf("%s %s, want: GET", req.Method, houseNumbersContextPath)
		t.FailNow()
	}

	_, _ = rw.Write([]byte(houseNumbersResponse))
}

func doGetServerRedirectUrl(t *testing.T, rw http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" && req.Method != "HEAD" {
		_ = fmt.Sprintf("%s %s, want: GET/HEAD", req.Method, RedirectUrlContextPath)
		t.FailNow()
	}

	if req.Method == "GET" || req.Method == "HEAD" {
		rw.Header().Add("Location", RedirectUrlHeader)
		_, _ = rw.Write([]byte(RedirectUrlResponse))
	}
}
