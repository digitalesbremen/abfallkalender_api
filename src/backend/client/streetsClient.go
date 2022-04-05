package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Streets []string

func (c *Client) GetStreets(redirectUrl string) (response Streets, err error) {
	url := fmt.Sprintf("%s%s", redirectUrl, "/Data/Strassen")
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

	s, err := io.ReadAll(resp.Body)

	_ = resp.Body.Close()

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
