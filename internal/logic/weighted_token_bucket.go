package logic

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const weightedTokenBucketPrefix string = "[weighted_token_bucket] "

type WeightedTokenBucket struct {
	sync.RWMutex
	sync.WaitGroup
	stopper           chan struct{}
	buckets           map[string]*atomic.Int64
	maxTokens         int64
	weightMultiplier  int64
	replinishInterval time.Duration
}

func NewWeightedTokenBucket(maxTokens, weightMultiplier int64, replinishInterval time.Duration) *WeightedTokenBucket {
	t := &WeightedTokenBucket{
		buckets:           make(map[string]*atomic.Int64),
		stopper:           make(chan struct{}),
		maxTokens:         maxTokens,
		replinishInterval: replinishInterval,
		weightMultiplier:  weightMultiplier,
	}
	t.launchReplinish()
	return t
}

func (t *WeightedTokenBucket) readBucket(id string) *atomic.Int64 {
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

func (t *WeightedTokenBucket) replinish() {
	t.Lock()
	defer t.Unlock()

	for _, i := range t.buckets {
		t := t.maxTokens
		i.Swap(t)
	}
	// fmt.Println(weightedTokenBucketPrefix + "tokens replenished")
}

func (t *WeightedTokenBucket) launchReplinish() {
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

func (t *WeightedTokenBucket) Limit(id string, weight int64) bool {
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

func (t *WeightedTokenBucket) Stop() {
	t.Lock()
	defer t.Unlock()

	close(t.stopper)
	t.Wait()
	fmt.Println(weightedTokenBucketPrefix + "stopped")
}
