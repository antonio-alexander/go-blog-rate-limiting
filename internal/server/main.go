package server

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/antonio-alexander/go-blog-rate-limiting/internal/logic"

	"github.com/pkg/errors"
)

func Main(pwd string, args []string, envs map[string]string, osSignal chan os.Signal) error {
	var wg sync.WaitGroup

	//get configuration
	config := NewConfiguration()
	config.FromEnvs(envs)

	//create server
	server := new(http.Server)
	server.Addr = ":" + config.Port

	//set rate limiter
	switch config.Algorithm {
	default:
		return errors.Errorf("unsupported algorithm: %s", config.Algorithm)
	case "":
		http.HandleFunc("/wait", endpointWait)
	case "weighted_token_bucket":
		rateLimiter := logic.NewWeightedTokenBucket(config.Maxtokens,
			config.WeightMultiplier, config.TokenReplinish)
		defer rateLimiter.Stop()
		http.HandleFunc("/wait", endpointWaitWeightedTokenBucket(rateLimiter))
	case "leaky_bucket":
		rateLimiter := logic.NewLeakyBucket(config.QueueSize, config.LeakRate)
		defer rateLimiter.Stop()
		http.HandleFunc("/wait", endpointWaitLeakyBucket(rateLimiter))
	case "token_bucket":
		rateLimiter := logic.NewTokenBucket(config.Maxtokens, config.TokenReplinish)
		defer rateLimiter.Stop()
		http.HandleFunc("/wait", endpointWaitTokenBucket(rateLimiter))
	}
	switch config.Algorithm {
	default:
		fmt.Printf(serverLogPrefix+"using algorithm: %s\n", config.Algorithm)
	case "":
		fmt.Printf(serverLogPrefix + "no algorithm configured")
	}

	//start server
	fmt.Printf(serverLogPrefix+"starting web server on %s:%s\n", config.Address, config.Port)
	chErr := make(chan error, 1)
	defer close(chErr)
	stopped := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(stopped)

		if err := server.ListenAndServe(); err != nil {
			chErr <- err
		}
	}()
	select {
	case <-stopped:
	case err := <-chErr:
		return err
	case <-osSignal:
		return server.Close()
	}
	wg.Wait()
	return nil
}
