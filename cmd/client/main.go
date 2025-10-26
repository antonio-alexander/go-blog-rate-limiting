package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/antonio-alexander/go-blog-rate-limiting/internal/client"
	"github.com/antonio-alexander/go-blog-rate-limiting/internal/config"
	"github.com/antonio-alexander/go-blog-rate-limiting/internal/data"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func generateId() string {
	return uuid.Must(uuid.NewRandom()).String()
}

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
	var wg sync.WaitGroup

	//get configuration
	config := config.NewConfiguration()
	config.FromCli(args)
	config.FromEnvs(envs)

	//create client
	client := client.New(config)

	//generate payload
	request := data.NewRequest()
	request.FromArgs(args)
	request.FromEnvs(envs)

	//generate context
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	go func() {
		defer wg.Done()
		defer cancel()

		close(started)
		<-osSignal
	}()
	<-started
	switch config.Mode {
	default:
		return errors.Errorf("unsupported mode: %s", config.Mode)
	case "single_request":
		//create client and execute request
		response, err := client.Wait(ctx, request)
		if err != nil {
			return err
		}
		bytes, err := json.Marshal(response)
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
	case "multiple_requests":
		//execute requests simultaneously
		start := make(chan struct{})
		for i := 0; i < config.NumberOfApplications; i++ {
			request.ApplicationId = generateId()
			for i := 0; i < config.NumberOfRequests; i++ {
				wg.Add(1)
				request.Id = generateId()
				go func(request data.Request) {
					defer wg.Done()

					tRequest := time.NewTicker(config.RequestRate)
					defer tRequest.Stop()
					<-start
					for {
						select {
						case <-osSignal:
							return
						case <-tRequest.C:
							tNow := time.Now()
							if _, err := client.Wait(ctx, &request); err != nil {
								fmt.Printf("%s(%s) error: %s\n", request.Id,
									request.ApplicationId, err.Error())
							}
							fmt.Printf("%s(%s): %v\n", request.Id,
								request.ApplicationId, time.Since(tNow))
						}
					}
				}(*request)
			}
		}
		close(start)
		wg.Wait()
	}
	return nil
}
