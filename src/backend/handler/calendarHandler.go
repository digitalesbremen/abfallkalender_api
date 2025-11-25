package handler

import (
	"encoding/json"
	"html"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

func (c Controller) GetCalendar(w http.ResponseWriter, r *http.Request) {
	streetName := parseStreetName(r)
	houseNumber := parseHouseNumber(r)

	redirectUrl, err := c.Client.GetRedirectUrl(InitialContextPath)

	if err != nil {
		// TODO handle 404
		c.createInternalServerError(w, err)
		return
	}

	// Content negotiation
	switch accept := getAcceptHeader(r); accept {
	case ICS:
		fallthrough
	case CSV:
		// Return calendar payload (ICS/CSV)
		var response []byte
		contentType := "text/calendar; charset=utf-8"
		if accept == CSV {
			response, err = c.Client.GetCSV(redirectUrl, url.QueryEscape(streetName), houseNumber)
			contentType = "text/csv; charset=utf-8"
		} else {
			response, err = c.Client.GetICS(redirectUrl, url.QueryEscape(streetName), houseNumber)
			contentType = "text/calendar; charset=utf-8"
		}
		if err != nil {
			c.createInternalServerError(w, err)
			return
		}
		if cacheStatus := c.Client.GetLastCacheStatus(); cacheStatus != "" {
			w.Header().Set("X-Cache", cacheStatus)
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(response)
		return
	case HTML:
		// Browser requested HTML → render a lightweight HTML preview instead of downloading ICS
		ics, err := c.Client.GetICS(redirectUrl, url.QueryEscape(streetName), houseNumber)
		if err != nil {
			c.createInternalServerError(w, err)
			return
		}
		// Simple HTML page that shows the ICS content inline to avoid a download experience in browsers
		// Note: We intentionally do NOT offer content negotiation here; this is a preview only.
		// If a client needs the raw ICS/CSV, it should set the appropriate Accept header.
		baseCalURL := buildHouseNumberUrl(r, streetName, houseNumber)
		// Build a webcal:// URL so users can subscribe in calendar apps directly.
		// Most OS/browser combinations hand off webcal:// links to the default calendar application.
		// We derive it from the absolute ICS URL by switching the scheme to "webcal".
		if u, perr := url.Parse(baseCalURL); perr == nil {
			u.Scheme = "webcal"
			baseCalURLWebcal := u.String()
			html := `<!doctype html><html lang="de"><head><meta charset="utf-8"><title>Abfallkalender – ` +
				html.EscapeString(streetName) + ` ` + html.EscapeString(houseNumber) + `</title>
            <meta name="viewport" content="width=device-width, initial-scale=1">
            <style>body{font-family:system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;margin:1.5rem;}
            pre{background:#f6f8fa;padding:1rem;overflow:auto;border-radius:6px;border:1px solid #e1e4e8;}
            a{color:#0366d6;text-decoration:none}a:hover{text-decoration:underline}
            .links{margin:0 0 1rem 0}
            .btn{display:inline-block;margin-left:1rem;padding:0.4rem 0.6rem;border:1px solid #0366d6;border-radius:4px}
            </style></head><body>
            <h1>Abfallkalender – ` + html.EscapeString(streetName) + ` ` + html.EscapeString(houseNumber) + `</h1>
            <div class="links">
              <a href="` + baseCalURL + `/next">Nächste Abholung (JSON)</a>
              <a class="btn" href="` + baseCalURLWebcal + `">In Kalender abonnieren (ICS)</a>
            </div>
            <p>Vorschau der ICS-Inhalte (nur Anzeige, kein Download):</p>
            <pre>` + html.EscapeString(string(ics)) + `</pre>
            </body></html>`
			if cacheStatus := c.Client.GetLastCacheStatus(); cacheStatus != "" {
				w.Header().Set("X-Cache", cacheStatus)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(html))
			return
		}
		// Fallback: if URL parsing failed for some reason, render without the subscribe button
		htmlStr := `<!doctype html><html lang="de"><head><meta charset="utf-8"><title>Abfallkalender – ` +
			html.EscapeString(streetName) + ` ` + html.EscapeString(houseNumber) + `</title>
            <meta name="viewport" content="width=device-width, initial-scale=1">
            <style>body{font-family:system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;margin:1.5rem;}
            pre{background:#f6f8fa;padding:1rem;overflow:auto;border-radius:6px;border:1px solid #e1e4e8;}
            a{color:#0366d6;text-decoration:none}a:hover{text-decoration:underline}
            .links{margin:0 0 1rem 0}
            </style></head><body>
            <h1>Abfallkalender – ` + html.EscapeString(streetName) + ` ` + html.EscapeString(houseNumber) + `</h1>
            <div class="links">
              <a href="` + baseCalURL + `/next">Nächste Abholung (JSON)</a>
            </div>
            <p>Vorschau der ICS-Inhalte (nur Anzeige, kein Download):</p>
            <pre>` + html.EscapeString(string(ics)) + `</pre>
            </body></html>`
		if cacheStatus := c.Client.GetLastCacheStatus(); cacheStatus != "" {
			w.Header().Set("X-Cache", cacheStatus)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(htmlStr))
		return
	case NONE:
		// No Accept header → keep legacy default (ICS) for CLI/cURL compatibility
		response, err := c.Client.GetICS(redirectUrl, url.QueryEscape(streetName), houseNumber)
		if err != nil {
			c.createInternalServerError(w, err)
			return
		}
		if cacheStatus := c.Client.GetLastCacheStatus(); cacheStatus != "" {
			w.Header().Set("X-Cache", cacheStatus)
		}
		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(response)
		return
	case JSON_ACCEPT:
		// Only when the client explicitly asks for JSON
		dto := buildHouseNumberResource(r, streetName, houseNumber)
		payload, err := json.Marshal(dto)
		if err != nil {
			c.createInternalServerError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
		return
	}
}

func parseHouseNumber(r *http.Request) string {
	params := mux.Vars(r)
	return params["number"]
}

func getAcceptHeader(r *http.Request) acceptHeader {
	if len(r.Header.Get("accept")) > 0 {
		for _, accept := range r.Header.Values("accept") {
			a := strings.ToLower(accept)
			if strings.Contains(a, "application/json") {
				return JSON_ACCEPT
			}
			if strings.Contains(a, "text/html") {
				return HTML
			}
			if strings.Contains(a, "text/calendar") {
				return ICS
			}
			if strings.Contains(a, "text/csv") {
				return CSV
			}
		}
	}
	return NONE
}

type acceptHeader int

const (
	NONE acceptHeader = iota
	ICS
	CSV
	JSON_ACCEPT
	HTML
)

// buildHouseNumberResource creates a minimal JSON representation for a single
// house number resource, including helpful links to calendar (ICS/CSV) and next.
func buildHouseNumberResource(r *http.Request, streetName string, houseNumber string) houseNumberDto {
	baseCalURL := buildHouseNumberUrl(r, streetName, houseNumber)
	dto := houseNumberDto{}
	dto.Number = houseNumber
	dto.Links.Self.Href = baseCalURL // self remains dereferenceable and will return ICS by default in browsers
	dto.Links.Next.Href = baseCalURL + "/next"
	return dto
}
