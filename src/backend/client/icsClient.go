package client

import (
	"fmt"
	"io"
	"net/http"
)

func (c *Client) GetICS(redirectUrl string, streetName string, houseNumber string) (response string, err error) {
	request, err := http.NewRequest("GET", buildICSUrl(redirectUrl, streetName, houseNumber), nil)

	// TODO make it cleaner! command - if err - command - if err - command if err?
	if err != nil {
		return "", err
	}

	resp, err := c.sendRequest(request, false)

	if err != nil {
		return "", err
	}

	ical, err := io.ReadAll(resp.Body)

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if err != nil {
		return "", err
	}

	return string(ical), nil
}

func buildICSUrl(redirectUrl string, streetName string, houseNumber string) string {
	return fmt.Sprintf("%s%s%s%s%s", redirectUrl, "/cal?strasse=", streetName, "&Hausnr=", houseNumber)
}
