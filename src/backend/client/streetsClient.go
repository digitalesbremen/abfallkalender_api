package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Streets []string

func (c *Client) GetStreets(redirectUrl string) (response Streets, err error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", redirectUrl, "/Data/Strassen"), nil)

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

	streets := make(Streets, 0)
	err = json.Unmarshal(s, &streets)

	if err != nil {
		return nil, err
	}

	// TODO trim values?
	streets.deleteEmptyStreets()

	return streets, nil
}

func (l *Streets) deleteEmptyStreets() {
	var r []string
	for _, str := range *l {
		if str != "" {
			r = append(r, str)
		}
	}
	*l = r
}
