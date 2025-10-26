package limiter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/antonio-alexander/go-blog-rate-limiting/internal/config"
	"github.com/antonio-alexander/go-blog-rate-limiting/internal/data"
)

const tokenBucketPrefix string = "[token_bucket] "

type tokenBucket struct {
	sync.RWMutex
	sync.WaitGroup
	config struct {
		maxTokens              int64
		tokenReplinishInterval time.Duration
	}
	stopper chan struct{}
	buckets map[string]*atomic.Int64
}

func NewToken(parameters ...any) Limiter {
	t := &tokenBucket{
		buckets: make(map[string]*atomic.Int64),
		stopper: make(chan struct{}),
	}
	for _, parameter := range parameters {
		switch p := parameter.(type) {
		case *config.Configuration:
			t.config.maxTokens = p.Maxtokens
			t.config.tokenReplinishInterval = p.TokenReplinish
		}
	}
	t.launchReplinish()
	return t
}

func (t *tokenBucket) readBucket(id string) *atomic.Int64 {
	t.RLock()
	defer t.RUnlock()

	i, ok := t.buckets[id]
	if !ok {
		i = new(atomic.Int64)
		i.Add(t.config.maxTokens)
		t.buckets[id] = i
	}
	return i
}

func (t *tokenBucket) replinish() {
	t.Lock()
	defer t.Unlock()

	for _, i := range t.buckets {
		t := t.config.maxTokens
		i.Swap(t)
	}
	fmt.Println(tokenBucketPrefix + "tokens replenished")
}

func (t *tokenBucket) launchReplinish() {
	t.Add(1)
	started := make(chan struct{})
	go func() {
		defer t.Done()

		tReplinish := time.NewTicker(t.config.tokenReplinishInterval)
		defer tReplinish.Stop()
		close(started)
		for {
			select {
			case <-t.stopper:
				return
			case <-tReplinish.C:
				t.replinish()
			}
		}
	}()
	<-started
}

func (t *tokenBucket) Limit(ctx context.Context, id string, parameters ...any) bool {
	i := t.readBucket(id)
	v := i.Load()
	if v <= 0 {
		fmt.Printf(tokenBucketPrefix+"%s limited (%d)\n", id, v)
		return true
	}
	v = i.Add(-1)
	fmt.Printf(tokenBucketPrefix+"%s allowed (%d)\n", id, v)
	return false
}

func (t *tokenBucket) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//read the request from bytes
		request := data.NewRequest()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf(tokenBucketPrefix+"error while writing bytes: %s\n", err.Error())
			}
			return
		}
		_ = r.Body.Close()
		if err := request.UnmarshalBinary(bodyBytes); err != nil {
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf(tokenBucketPrefix+"error while writing bytes: %s\n", err.Error())
			}
			return
		}

		//execute the rate limiter
		if t.Limit(r.Context(), request.ApplicationId, int64(request.Weight)) {
			retryAfter := t.config.tokenReplinishInterval.Seconds() //this isn't going to be consistent
			bytes := []byte("too many requests received")
			w.Header().Add("Retry-After", fmt.Sprint(retryAfter))
			w.Header().Set("Content-Length", fmt.Sprint(len(bytes)))
			w.WriteHeader(http.StatusTooManyRequests)
			if _, err := w.Write(bytes); err != nil {
				fmt.Printf(tokenBucketPrefix+"error while writing bytes: %s\n", err.Error())
			}
			return
		}

		//execute next endpoint
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		next(w, r)
	})
}

func (t *tokenBucket) Stop() {
	t.Lock()
	defer t.Unlock()

	close(t.stopper)
	t.Wait()
	fmt.Println(tokenBucketPrefix + "stopped")
}
