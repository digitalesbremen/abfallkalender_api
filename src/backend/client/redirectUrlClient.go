package client

import (
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) GetRedirectUrl(contextPath string) (url string, err error) {
	req, err := http.NewRequest("HEAD", fmt.Sprintf("%s%s", c.BaseURL, contextPath), nil)

	if err != nil {
		return "", err
	}

	response, err := c.sendRequest(req)
	if err := err; err != nil {
		return "", err
	}

	return c.BaseURL + strings.Replace(response.Header.Get("Location"), "/Abfallkalender", "", 1), nil
}
