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

const weightedTokenBucketPrefix string = "[weighted_token_bucket] "

type weightedTokenBucket struct {
	sync.RWMutex
	sync.WaitGroup
	stopper           chan struct{}
	buckets           map[string]*atomic.Int64
	maxTokens         int64
	weightMultiplier  int64
	replinishInterval time.Duration
}

func NewWeighted(parameters ...any) Limiter {
	t := &weightedTokenBucket{
		buckets: make(map[string]*atomic.Int64),
		stopper: make(chan struct{}),
	}
	for _, parameter := range parameters {
		switch p := parameter.(type) {
		case *config.Configuration:
			t.maxTokens = p.Maxtokens
			t.replinishInterval = p.TokenReplinish
			t.weightMultiplier = p.WeightMultiplier
		}
	}
	t.launchReplinish()
	return t
}

func (t *weightedTokenBucket) readBucket(id string) *atomic.Int64 {
	t.RLock()
	defer t.RUnlock()
	i, ok := t.buckets[id]
	if !ok {
		i = new(atomic.Int64)
		i.Add(t.maxTokens)
		t.buckets[id] = i
	}
	return i
}

func (t *weightedTokenBucket) replinish() {
	t.Lock()
	defer t.Unlock()

	for _, i := range t.buckets {
		t := t.maxTokens
		i.Swap(t)
	}
	// fmt.Println(weightedTokenBucketPrefix + "tokens replenished")
}

func (t *weightedTokenBucket) launchReplinish() {
	t.Add(1)
	started := make(chan struct{})
	go func() {
		defer t.Done()

		tReplinish := time.NewTicker(t.replinishInterval)
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

func (t *weightedTokenBucket) Limit(ctx context.Context, id string, parameters ...any) bool {
	var weight int64

	for _, parameter := range parameters {
		if i, ok := parameter.(int64); ok {
			weight = i
		}
	}
	i := t.readBucket(id)
	v := i.Load()
	if v <= 0 {
		fmt.Printf(weightedTokenBucketPrefix+"%s limited (%d)\n", id, v)
		return true
	}
	v = i.Add(-1 * t.weightMultiplier * weight)
	fmt.Printf(weightedTokenBucketPrefix+"%s allowed (%d)\n", id, v)
	return false
}

func (t *weightedTokenBucket) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//read the request from bytes
		request := data.NewRequest()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf(weightedTokenBucketPrefix+"error while writing bytes: %s\n", err.Error())
			}
			return
		}
		_ = r.Body.Close()
		if err := request.UnmarshalBinary(bodyBytes); err != nil {
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf(weightedTokenBucketPrefix+"error while writing bytes: %s\n", err.Error())
			}
			return
		}

		//execute the rate limiter
		if t.Limit(r.Context(), request.ApplicationId, int64(request.Weight)) {
			bytes := []byte("too many requests received")
			w.Header().Set("Content-Length", fmt.Sprint(len(bytes)))
			w.WriteHeader(http.StatusTooManyRequests)
			if _, err := w.Write(bytes); err != nil {
				fmt.Printf(weightedTokenBucketPrefix+"error while writing bytes: %s\n", err.Error())
			}
			return
		}

		//execute next endpoint
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		next(w, r)
	})
}

func (t *weightedTokenBucket) Stop() {
	t.Lock()
	defer t.Unlock()

	close(t.stopper)
	t.Wait()
	fmt.Println(weightedTokenBucketPrefix + "stopped")
}
