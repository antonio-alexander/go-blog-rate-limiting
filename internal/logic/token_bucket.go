package logic

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const tokenBucketPrefix string = "[token_bucket] "

type TokenBucket struct {
	sync.RWMutex
	sync.WaitGroup
	stopper           chan struct{}
	buckets           map[string]*atomic.Int64
	maxTokens         int64
	replinishInterval time.Duration
}

func NewTokenBucket(maxTokens int64, replinishInterval time.Duration) *TokenBucket {
	t := &TokenBucket{
		buckets:           make(map[string]*atomic.Int64),
		stopper:           make(chan struct{}),
		maxTokens:         maxTokens,
		replinishInterval: replinishInterval,
	}
	t.launchReplinish()
	return t
}

func (t *TokenBucket) readBucket(id string) *atomic.Int64 {
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

func (t *TokenBucket) replinish() {
	t.Lock()
	defer t.Unlock()

	for _, i := range t.buckets {
		t := t.maxTokens
		i.Swap(t)
	}
	// fmt.Println(tokenBucketPrefix + "tokens replenished")
}

func (t *TokenBucket) launchReplinish() {
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

func (t *TokenBucket) Limit(id string) bool {
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

func (t *TokenBucket) Stop() {
	t.Lock()
	defer t.Unlock()

	close(t.stopper)
	t.Wait()
	fmt.Println(tokenBucketPrefix + "stopped")
}
