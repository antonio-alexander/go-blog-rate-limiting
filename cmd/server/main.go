package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/antonio-alexander/go-blog-rate-limiting/internal/config"
	"github.com/antonio-alexander/go-blog-rate-limiting/internal/limiter"
	"github.com/antonio-alexander/go-blog-rate-limiting/internal/server"

	"github.com/pkg/errors"
)

func main() {
	pwd, _ := os.Getwd()
	args := os.Args[1:]
	envs := make(map[string]string)
	for _, env := range os.Environ() {
		if s := strings.Split(env, "="); len(s) > 1 {
			switch {
			case len(s) == 2:
				envs[s[0]] = s[1]
			case len(s) > 2:
				envs[s[0]] = strings.Join(s[1:], "=")
			}
		}
	}
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)
	if err := Main(pwd, args, envs, osSignal); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}

func Main(pwd string, args []string, envs map[string]string, osSignal chan os.Signal) error {
	var rateLimiter limiter.Limiter

	//get configuration
	config := config.NewConfiguration()
	config.FromEnvs(envs)

	//create rate limiter if configured
	switch limiter.LimiterType(config.Algorithm) {
	default:
		return errors.Errorf("unsupported algorithm: %s", config.Algorithm)
	case limiter.LimiterTypeWeighted:
		rateLimiter = limiter.NewWeighted(config)
	case limiter.LimiterTypeLeaky:
		rateLimiter = limiter.NewLeaky(config)
	case limiter.LimiterTypeToken:
		rateLimiter = limiter.NewToken(config)
	}
	fmt.Printf("configured rate limiting algorithim: %s\n", config.Algorithm)
	if rateLimiter != nil {
		defer rateLimiter.Stop()
	}

	//create and start server
	server := server.New(config, rateLimiter)
	if err := server.Start(); err != nil {
		return err
	}
	<-osSignal
	return server.Stop()
}
