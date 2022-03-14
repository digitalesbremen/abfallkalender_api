package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) sendRequest(originalRequest *http.Request, autoCloseBody bool) (*http.Response, error) {
	request := originalRequest
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Accept", "application/json; charset=utf-8")

	res, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}

	if autoCloseBody {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return nil, errors.New(errRes.Message)
		}

		return nil, fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	return res, nil
}
