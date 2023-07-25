package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/antonio-alexander/go-blog-rate-limiting/internal/data"

	"github.com/pkg/errors"
)

// const clientLogPrefix string = "[client] "

type client struct {
	address    string
	timeout    time.Duration
	retry      bool
	maxRetries int
}

func New(c *Configuration) interface {
	Wait(context.Context, *data.Request) (*data.Response, error)
} {
	address := "http://" + c.Address
	if c.Port != "" {
		address = address + ":" + c.Port
	}
	return &client{
		address: address,
		timeout: c.Timeout,
		retry:   c.Retry,
	}
}

func (c *client) doRequest(ctx context.Context, uri, method string, data []byte) ([]byte, int, error) {
	client := new(http.Client)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, method, uri, bytes.NewBuffer(data))
	if err != nil {
		return nil, -1, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, -1, err
	}
	if c.retry && response.StatusCode == http.StatusTooManyRequests {
		i, _ := strconv.ParseInt(response.Header.Get("Retry-After"), 10, 64)
		if i <= 0 {
			i = 1000
		}
		tRetry := time.NewTicker(time.Millisecond * time.Duration(i))
		defer tRetry.Stop()
		for i := 0; i < c.maxRetries; i++ {
			<-tRetry.C
			response, err = client.Do(request)
			if err != nil {
				return nil, -1, err
			}
			if response.StatusCode != http.StatusTooManyRequests {
				break
			}
		}
	}
	data, err = io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, -1, err
	}
	return data, response.StatusCode, nil
}

func (c *client) Wait(ctx context.Context, request *data.Request) (*data.Response, error) {
	//generate payload and uri
	bytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	uri := c.address + "/wait"

	//execute request
	bytes, statusCode, err := c.doRequest(ctx, uri, http.MethodPost, bytes)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", statusCode)
	}
	response := &data.Response{}
	if err := json.Unmarshal(bytes, response); err != nil {
		return nil, err
	}
	return response, nil
}
