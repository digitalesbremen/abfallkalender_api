package client

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func (c *Client) GetICS(redirectUrl string, streetName string, houseNumber string) (response []byte, err error) {
	url := buildICSUrl(redirectUrl, streetName, houseNumber)
	request, err := http.NewRequest("GET", url, nil)

	log.Printf("Call URL '%s'\n", url)

	// TODO make it cleaner! command - if err - command - if err - command if err?
	if err != nil {
		return nil, err
	}

	resp, err := c.sendRequest(request, false)

	if err != nil {
		return nil, err
	}

	ical, err := io.ReadAll(resp.Body)

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if err != nil {
		return nil, err
	}

	return ical, nil
}

func buildICSUrl(redirectUrl string, streetName string, houseNumber string) string {
	return fmt.Sprintf("%s%s%s%s%s", redirectUrl, "/Abfallkalender/cal?strasse=", streetName, "&Hausnr=", houseNumber)
}
