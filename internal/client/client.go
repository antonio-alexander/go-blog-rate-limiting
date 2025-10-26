package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/antonio-alexander/go-blog-rate-limiting/internal/config"
	"github.com/antonio-alexander/go-blog-rate-limiting/internal/data"

	"github.com/pkg/errors"
)

// const clientLogPrefix string = "[client] "

type client struct {
	config struct {
		timeout    time.Duration
		retry      bool
		maxRetries int
	}
	address string
}

type Client interface {
	Wait(context.Context, *data.Request) (*data.Response, error)
}

func New(parameters ...any) Client {
	c := &client{}
	for _, parameter := range parameters {
		switch p := parameter.(type) {
		case *config.Configuration:
			c.address = "http://" + p.Host
			if p.Port != "" {
				c.address += ":" + p.Port
			}
			c.config.timeout = p.Timeout
			c.config.retry = p.Retry
			c.config.maxRetries = p.MaxRetries
		}
	}
	return c
}

func (c *client) doRequest(ctx context.Context, uri, method string, data []byte) ([]byte, int, error) {
	client := new(http.Client)
	ctx, cancel := context.WithTimeout(ctx, c.config.timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, method, uri, bytes.NewBuffer(data))
	if err != nil {
		return nil, -1, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, -1, err
	}
	if c.config.retry && response.StatusCode == http.StatusTooManyRequests {
		i, _ := strconv.ParseInt(response.Header.Get("Retry-After"), 10, 64)
		if i <= 0 {
			i = 1000
		}
		tRetry := time.NewTicker(time.Millisecond * time.Duration(i))
		defer tRetry.Stop()
		for i := 0; i < c.config.maxRetries; i++ {
			<-tRetry.C
			ctx, cancel := context.WithTimeout(ctx, c.config.timeout)
			defer cancel()
			request = request.WithContext(ctx)
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
	if err != nil {
		return nil, -1, err
	}
	response.Body.Close()
	return data, response.StatusCode, nil
}

func (c *client) Wait(ctx context.Context, request *data.Request) (*data.Response, error) {
	//generate payload and uri
	bytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	uri := c.address + data.RouteWait

	//execute request
	bytes, statusCode, err := c.doRequest(ctx, uri, data.MethodWait, bytes)
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
