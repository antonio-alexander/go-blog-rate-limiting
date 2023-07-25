package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/antonio-alexander/go-blog-rate-limiting/internal/data"
	"github.com/antonio-alexander/go-blog-rate-limiting/internal/logic"
	"github.com/pkg/errors"
)

const (
	RETRY_AFTER     string = "Retry-After"
	serverLogPrefix string = "[server] "
)

func endpointWait(w http.ResponseWriter, r *http.Request) {
	request := data.NewRequest()
	if err := request.FromRequest(r); err != nil {
		fmt.Printf("error while reading body: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = w.Write([]byte(err.Error())); err != nil {
			fmt.Printf("error while writing bytes: %s\n", err.Error())
		}
		return
	}

	//execute business logic
	select {
	case <-time.After(request.Wait):
		fmt.Printf("%s: wait of %v completed\n", request.Id, request.Wait)
	case <-r.Context().Done():
		fmt.Printf("%s: wait cancelled (context)\n", request.Id)
		return
	}

	//marshal response
	bytes, err := json.Marshal(&data.Response{
		Id:     request.Id,
		Wait:   request.Wait,
		Weight: request.Weight,
	})
	if err != nil {
		fmt.Printf("error while marshalling response: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = w.Write([]byte(err.Error())); err != nil {
			fmt.Printf("error while writing bytes: %s\n", err.Error())
		}
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprint(len(bytes)))
	if _, err = w.Write(bytes); err != nil {
		fmt.Printf("error while writing bytes: %s\n", err.Error())
	}
}

func endpointWaitWeightedTokenBucket(rateLimiter *logic.WeightedTokenBucket) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		request := data.NewRequest()
		if err := request.FromRequest(r); err != nil {
			fmt.Printf("error while reading body: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf("error while writing bytes: %s\n", err.Error())
			}
			return
		}

		//execute the rate limiter
		if rateLimiter.Limit(request.ApplicationId, int64(request.Weight)) {
			err := errors.Errorf("too many requests received")
			w.WriteHeader(http.StatusTooManyRequests)
			//REVIEW: would like for this to be dynamic
			w.Header().Add(RETRY_AFTER, "3600")
			if _, err := w.Write([]byte(err.Error())); err != nil {
				fmt.Printf("error while writing bytes: %s\n", err.Error())
			}
			return
		}

		//execute business logic
		select {
		case <-time.After(request.Wait):
			fmt.Printf("%s: wait of %v completed\n", request.Id, request.Wait)
		case <-r.Context().Done():
			fmt.Printf("%s: wait cancelled (context)\n", request.Id)
			return
		}

		//marshal response
		bytes, err := json.Marshal(&data.Response{
			Id:     request.Id,
			Wait:   request.Wait,
			Weight: request.Weight,
		})
		if err != nil {
			fmt.Printf("error while marshalling response: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf("error while writing bytes: %s\n", err.Error())
			}
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Length", fmt.Sprint(len(bytes)))
		if _, err = w.Write(bytes); err != nil {
			fmt.Printf("error while writing bytes: %s\n", err.Error())
		}
	}
}

func endpointWaitTokenBucket(rateLimiter *logic.TokenBucket) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		request := data.NewRequest()
		if err := request.FromRequest(r); err != nil {
			fmt.Printf("error while reading body: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf("error while writing bytes: %s\n", err.Error())
			}
			return
		}

		//execute the rate limiter
		if rateLimiter.Limit(request.ApplicationId) {
			err := errors.Errorf("too many requests received")
			w.WriteHeader(http.StatusTooManyRequests)
			//REVIEW: would like for this to be dynamic
			w.Header().Add(RETRY_AFTER, "3600")
			if _, err := w.Write([]byte(err.Error())); err != nil {
				fmt.Printf("error while writing bytes: %s\n", err.Error())
			}
			return
		}

		//execute business logic
		select {
		case <-time.After(request.Wait):
			fmt.Printf("%s: wait of %v completed\n", request.Id, request.Wait)
		case <-r.Context().Done():
			fmt.Printf("%s: wait cancelled (context)\n", request.Id)
			return
		}

		//marshal response
		bytes, err := json.Marshal(&data.Response{
			Id:     request.Id,
			Wait:   request.Wait,
			Weight: request.Weight,
		})
		if err != nil {
			fmt.Printf("error while marshalling response: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf("error while writing bytes: %s\n", err.Error())
			}
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Length", fmt.Sprint(len(bytes)))
		if _, err = w.Write(bytes); err != nil {
			fmt.Printf("error while writing bytes: %s\n", err.Error())
		}
	}
}

func endpointWaitLeakyBucket(rateLimiter *logic.LeakyBucket) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		request := data.NewRequest()
		if err := request.FromRequest(r); err != nil {
			fmt.Printf("error while reading body: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf("error while writing bytes: %s\n", err.Error())
			}
			return
		}

		//execute the rate limiter
		limited, signalOut := rateLimiter.Limit(request.ApplicationId)
		if limited {
			err := errors.Errorf("too many requests received")
			w.WriteHeader(http.StatusTooManyRequests)
			//REVIEW: would like for this to be dynamic
			w.Header().Add(RETRY_AFTER, "3600")
			if _, err := w.Write([]byte(err.Error())); err != nil {
				fmt.Printf("error while writing bytes: %s\n", err.Error())
			}
			return
		}

		select {
		case <-r.Context().Done():
			fmt.Printf("%s: wait cancelled (context)\n", request.Id)
			return
		case <-signalOut:
		}

		//execute business logic
		select {
		case <-r.Context().Done():
			fmt.Printf("%s: wait cancelled (context)\n", request.Id)
			return
		case <-time.After(request.Wait):
			fmt.Printf("%s: wait of %v completed\n", request.Id, request.Wait)
		}

		//marshal response
		bytes, err := json.Marshal(&data.Response{
			Id:     request.Id,
			Wait:   request.Wait,
			Weight: request.Weight,
		})
		if err != nil {
			fmt.Printf("error while marshalling response: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf("error while writing bytes: %s\n", err.Error())
			}
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Length", fmt.Sprint(len(bytes)))
		if _, err = w.Write(bytes); err != nil {
			fmt.Printf("error while writing bytes: %s\n", err.Error())
		}

	}
}
