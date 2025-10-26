package limiter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/antonio-alexander/go-blog-rate-limiting/internal/config"
	"github.com/antonio-alexander/go-blog-rate-limiting/internal/data"

	goqueue "github.com/antonio-alexander/go-queue"
	"github.com/antonio-alexander/go-queue/finite"
)

const leakyBucketPrefix string = "[leaky_bucket] "

type queue interface {
	goqueue.Owner
	goqueue.GarbageCollecter
	goqueue.Dequeuer
	goqueue.Enqueuer
	goqueue.EnqueueInFronter
	goqueue.Length
	goqueue.Event
	goqueue.Peeker
	finite.EnqueueLossy
	finite.Resizer
	finite.Capacity
}

type leakyBucket struct {
	sync.RWMutex
	sync.WaitGroup
	stopper   chan struct{}
	buckets   map[string]queue
	queueSize int
	leakRate  time.Duration
}

func NewLeaky(parameters ...any) Limiter {
	l := &leakyBucket{
		buckets: make(map[string]queue),
		stopper: make(chan struct{}),
	}
	for _, parameter := range parameters {
		switch p := parameter.(type) {
		case *config.Configuration:
			l.queueSize = p.QueueSize
			l.leakRate = p.LeakRate
		}
	}
	return l
}

func (l *leakyBucket) launchHandler(q queue) {
	l.Add(1)
	started := make(chan struct{})
	go func() {
		defer l.Done()

		tLeakRate := time.NewTicker(l.leakRate)
		defer tLeakRate.Stop()
		close(started)
		for {
			select {
			case <-l.stopper:
				return
			case <-tLeakRate.C:
				q.Dequeue()
			}
		}
	}()
	<-started
}

func (l *leakyBucket) Limit(ctx context.Context, id string, parameters ...any) bool {
	l.Lock()
	defer l.Unlock()

	queue, ok := l.buckets[id]
	if !ok {
		queue = finite.New(l.queueSize)
		l.buckets[id] = queue
		l.launchHandler(queue)
	}
	if overflow := queue.Enqueue(struct{}{}); overflow {
		fmt.Printf(leakyBucketPrefix+"%s limited (%d)\n", id, queue.Length())
		for stop := false; stop; {
			tEnqueue := time.NewTicker(time.Millisecond)
			defer tEnqueue.Stop()
			select {
			case <-ctx.Done():
				return true
			case <-tEnqueue.C:
				if overflow := queue.Enqueue(struct{}{}); !overflow {
					stop = true
				}
			}
		}
	}
	fmt.Printf(leakyBucketPrefix+"%s allowed (%d)\n", id, queue.Length())
	return false
}

func (l *leakyBucket) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//read the request from bytes
		request := data.NewRequest()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf(leakyBucketPrefix+"error while writing bytes: %s\n", err.Error())
			}
			return
		}
		_ = r.Body.Close()
		if err := request.UnmarshalBinary(bodyBytes); err != nil {
			if _, err = w.Write([]byte(err.Error())); err != nil {
				fmt.Printf(leakyBucketPrefix+"error while writing bytes: %s\n", err.Error())
			}
			return
		}

		//execute the rate limiter
		if l.Limit(r.Context(), request.ApplicationId, int64(request.Weight)) {
			bytes := []byte("too many requests received")
			w.Header().Set("Content-Length", fmt.Sprint(len(bytes)))
			w.WriteHeader(http.StatusTooManyRequests)
			if _, err := w.Write(bytes); err != nil {
				fmt.Printf(leakyBucketPrefix+"error while writing bytes: %s\n", err.Error())
			}
			return
		}

		//execute next endpoint
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		next(w, r)
	})
}

func (l *leakyBucket) Stop() {
	l.Lock()
	defer l.Unlock()

	close(l.stopper)
	l.Wait()
	fmt.Println(leakyBucketPrefix + "stopped")
}
