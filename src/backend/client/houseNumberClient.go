package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HouseNumbers []string

func (c *Client) GetHouseNumbers(redirectUrl string, streetName string) (response HouseNumbers, err error) {
	request, err := http.NewRequest("GET", buildUrl(redirectUrl, streetName), nil)

	// TODO make it cleaner! command - if err - command - if err - command if err?
	if err != nil {
		return nil, err
	}

	resp, err := c.sendRequest(request, false)

	if err != nil {
		return nil, err
	}

	s, err := io.ReadAll(resp.Body)

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if err != nil {
		return nil, err
	}

	houseNumbers := make(HouseNumbers, 0)
	err = json.Unmarshal(s, &houseNumbers)

	if err != nil {
		return nil, err
	}

	// TODO trim values?
	houseNumbers.deleteEmptyStreets()

	return houseNumbers, nil
}

func buildUrl(redirectUrl string, streetName string) string {
	return fmt.Sprintf("%s%s%s", redirectUrl, "/Data/Hausnummern?strasse=", streetName)
}

func (l *HouseNumbers) deleteEmptyStreets() {
	var r []string
	for _, str := range *l {
		if str != "" {
			r = append(r, str)
		}
	}
	*l = r
}
