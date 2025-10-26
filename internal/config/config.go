package config

import (
	"flag"
	"strconv"
	"time"
)

const (
	DefaultHttpPort             string        = "8080"
	DefaultHost                 string        = "localhost"
	DefaultTimeout              time.Duration = 10 * time.Second
	DefaultRequestRate          time.Duration = time.Second
	DefaultMode                 string        = "single_request"
	DefaultNumberOfRequests     int           = 4
	DefaultNumberOfApplications int           = 1
	DefaultRetry                bool          = false
	DefaultMaxRetries           int           = 2
	DefaultMaxtokens            int64         = 4
	DefaultTokenReplinsh        time.Duration = time.Second
	DefaultLeakRate             time.Duration = 250 * time.Millisecond
	DefaultQueueSize            int           = 4
	DefaultWeightMultiplier     int64         = 1
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
	ALGORITHM              string = "ALGORITHM"
	MAX_TOKENS             string = "MAX_TOKENS"
	TOKEN_REPLENISH        string = "TOKEN_REPLENISH_S"
	QUEUE_SIZE             string = "QUEUE_SIZE"
	LEAK_RATE              string = "LEAK_RATE_S"
	WEIGHT_MULTIPLIER      string = "WEIGHT_MULTIPLIER"
)

type Configuration struct {
	Host                 string
	Port                 string
	Timeout              time.Duration
	Mode                 string
	NumberOfRequests     int
	NumberOfApplications int
	RequestRate          time.Duration
	Retry                bool
	MaxRetries           int
	Maxtokens            int64
	TokenReplinish       time.Duration
	Algorithm            string
	QueueSize            int
	LeakRate             time.Duration
	WeightMultiplier     int64
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Host:                 DefaultHost,
		Port:                 DefaultHttpPort,
		Timeout:              DefaultTimeout,
		Mode:                 DefaultMode,
		NumberOfRequests:     DefaultNumberOfRequests,
		NumberOfApplications: DefaultNumberOfApplications,
		RequestRate:          DefaultRequestRate,
		Retry:                DefaultRetry,
		MaxRetries:           DefaultMaxRetries,
		Maxtokens:            DefaultMaxtokens,
		TokenReplinish:       DefaultTokenReplinsh,
		QueueSize:            DefaultQueueSize,
		LeakRate:             DefaultLeakRate,
		WeightMultiplier:     DefaultWeightMultiplier,
	}
}

func (c *Configuration) FromEnvs(envs map[string]string) {
	if s := envs[HTTP_PORT]; s != "" {
		c.Port = s
	}
	if s := envs[HTTP_ADDRESS]; s != "" {
		c.Host = s
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
	if s := envs[ALGORITHM]; s != "" {
		c.Algorithm = s
	}
	if s := envs[MAX_TOKENS]; s != "" {
		c.Maxtokens, _ = strconv.ParseInt(s, 10, 64)
	}
	if s := envs[QUEUE_SIZE]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		c.QueueSize = int(i)
	}
	if s := envs[TOKEN_REPLENISH]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		c.TokenReplinish = time.Duration(i) * time.Minute
	}
	if s := envs[LEAK_RATE]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		c.LeakRate = time.Duration(i) * time.Millisecond
	}
	if s := envs[WEIGHT_MULTIPLIER]; s != "" {
		c.WeightMultiplier, _ = strconv.ParseInt(s, 10, 64)
	}
}

func (c *Configuration) FromCli(args []string) {
	var timeout, requestRate int

	if len(args) <= 0 {
		return
	}
	flag.StringVar(&c.Host, "address", DefaultHost, "")
	flag.StringVar(&c.Port, "port", DefaultHttpPort, "")
	flag.StringVar(&c.Mode, "mode", DefaultMode, "")
	flag.IntVar(&c.NumberOfApplications, "number-of-applications", DefaultNumberOfApplications, "")
	flag.IntVar(&c.NumberOfRequests, "number-of-requests", DefaultNumberOfRequests, "")
	flag.IntVar(&timeout, "timeout", int(DefaultTimeout/time.Second), "")
	flag.IntVar(&requestRate, "request-rate", int(DefaultRequestRate/time.Second), "")
	flag.BoolVar(&c.Retry, "retry", DefaultRetry, "")
	flag.IntVar(&c.MaxRetries, "max-retries", DefaultMaxRetries, "")
	flag.Parse()
	c.Timeout = time.Second * time.Duration(timeout)
	c.RequestRate = time.Second * time.Duration(requestRate)
}
