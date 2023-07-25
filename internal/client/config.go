package client

import (
	"flag"
	"strconv"
	"time"
)

const (
	defaultHttpPort             string        = "8080"
	defaultAddress              string        = "localhost"
	defaultTimeout              time.Duration = time.Hour
	defaultRequestRate          time.Duration = time.Second
	defaultMode                 string        = "single_request"
	defaultNumberOfRequests     int           = 4
	defaultNumberOfApplications int           = 1
	defaultRetry                bool          = true
	defaultMaxRetries           int           = 2
)

const (
	HTTP_ADDRESS           string = "HTTP_ADDRESS"
	HTTP_PORT              string = "HTTP_PORT"
	TIMEOUT                string = "TIMEOUT"
	MODE                   string = "MODE"
	NUMBER_OF_REQUESTS     string = "NUMBER_OF_REQUESTS"
	NUMBER_OF_APPLICATIONS string = "NUMBER_OF_APPLICATIONS"
	REQUEST_RATE           string = "REQUEST_RATE"
	RETRY                  string = "RETRY"
	MAX_RETRIES            string = "MAX_RETRIES"
)

type Configuration struct {
	Address              string
	Port                 string
	Timeout              time.Duration
	Mode                 string
	NumberOfRequests     int
	NumberOfApplications int
	RequestRate          time.Duration
	Retry                bool
	MaxRetries           int
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Address:              defaultAddress,
		Port:                 defaultHttpPort,
		Timeout:              defaultTimeout,
		Mode:                 defaultMode,
		NumberOfRequests:     defaultNumberOfRequests,
		NumberOfApplications: defaultNumberOfApplications,
		RequestRate:          defaultRequestRate,
		Retry:                defaultRetry,
		MaxRetries:           defaultMaxRetries,
	}
}

func (c *Configuration) FromEnvs(envs map[string]string) {
	if s := envs[HTTP_PORT]; s != "" {
		c.Port = s
	}
	if s := envs[HTTP_ADDRESS]; s != "" {
		c.Address = s
	}
	if s := envs[TIMEOUT]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		c.Timeout = time.Duration(i) * time.Second
	}
	if s := envs[REQUEST_RATE]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		c.RequestRate = time.Duration(i) * time.Second
	}
	if s := envs[MODE]; s != "" {
		c.Mode = s
	}
	if s := envs[NUMBER_OF_REQUESTS]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		c.NumberOfRequests = int(i)
	}
	if s := envs[NUMBER_OF_APPLICATIONS]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		c.NumberOfApplications = int(i)
	}
	if s := envs[RETRY]; s != "" {
		c.Retry, _ = strconv.ParseBool(s)
	}
	if s := envs[MAX_RETRIES]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		c.MaxRetries = int(i)
	}
}

func (c *Configuration) FromCli(args []string) {
	var timeout, requestRate int

	if len(args) <= 0 {
		return
	}
	flag.StringVar(&c.Address, "address", defaultAddress, "")
	flag.StringVar(&c.Port, "port", defaultHttpPort, "")
	flag.StringVar(&c.Mode, "mode", defaultMode, "")
	flag.IntVar(&c.NumberOfApplications, "number-of-applications", defaultNumberOfApplications, "")
	flag.IntVar(&c.NumberOfRequests, "number-of-requests", defaultNumberOfRequests, "")
	flag.IntVar(&timeout, "timeout", int(defaultTimeout/time.Second), "")
	flag.IntVar(&requestRate, "request-rate", int(defaultRequestRate/time.Second), "")
	flag.BoolVar(&c.Retry, "retry", defaultRetry, "")
	flag.IntVar(&c.MaxRetries, "max-retries", defaultMaxRetries, "")
	flag.Parse()
	c.Timeout = time.Second * time.Duration(timeout)
	c.RequestRate = time.Second * time.Duration(requestRate)
}
