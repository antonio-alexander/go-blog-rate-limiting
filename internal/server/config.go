package server

import (
	"strconv"
	"time"
)

const (
	defaultAddress          string        = ""
	defaulHttpPort          string        = "8080"
	defaultReadTimeout      time.Duration = time.Minute
	defaultWriteTimeout     time.Duration = time.Minute
	defaultAlgorithm        string        = "token_bucket"
	defaultMaxtokens        int64         = 4
	defaultTokenReplinsh    time.Duration = time.Second
	defaultLeakRate         time.Duration = 250 * time.Millisecond
	defaultQueueSize        int           = 4
	defaultWeightMultiplier int64         = 1
)

const (
	HTTP_ADDRESS      string = "HTTP_ADDRESS"
	HTTP_PORT         string = "HTTP_PORT"
	READ_TIMEOUT      string = "READ_TIMEOUT"
	WRITE_TIMEOUT     string = "WRITE_TIMEOUT"
	ALGORITHM         string = "ALGORITHM"
	MAX_TOKENS        string = "MAX_TOKENS"
	TOKEN_REPLENISH   string = "TOKEN_REPLENISH"
	QUEUE_SIZE        string = "QUEUE_SIZE"
	LEAK_RATE         string = "LEAK_RATE"
	WEIGHT_MULTIPLIER string = "WEIGHT_MULTIPLIER"
)

type Configuration struct {
	Address          string
	Port             string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	Maxtokens        int64
	TokenReplinish   time.Duration
	Algorithm        string
	QueueSize        int
	LeakRate         time.Duration
	WeightMultiplier int64
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Address:          defaultAddress,
		Port:             defaulHttpPort,
		ReadTimeout:      defaultReadTimeout,
		WriteTimeout:     defaultWriteTimeout,
		Maxtokens:        defaultMaxtokens,
		TokenReplinish:   defaultTokenReplinsh,
		Algorithm:        defaultAlgorithm,
		QueueSize:        defaultQueueSize,
		LeakRate:         defaultLeakRate,
		WeightMultiplier: defaultWeightMultiplier,
	}
}

func (c *Configuration) FromEnvs(envs map[string]string) {
	if s := envs[HTTP_PORT]; s != "" {
		c.Port = s
	}
	if s := envs[HTTP_ADDRESS]; s != "" {
		c.Address = s
	}
	if s := envs[READ_TIMEOUT]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		c.ReadTimeout = time.Duration(i) * time.Minute
	}
	if s := envs[WRITE_TIMEOUT]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		c.WriteTimeout = time.Duration(i) * time.Minute
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
		c.WriteTimeout = time.Duration(i) * time.Minute
	}
	if s := envs[LEAK_RATE]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		c.LeakRate = time.Duration(i) * time.Millisecond
	}
	if s := envs[WEIGHT_MULTIPLIER]; s != "" {
		c.WeightMultiplier, _ = strconv.ParseInt(s, 10, 64)
	}
}
