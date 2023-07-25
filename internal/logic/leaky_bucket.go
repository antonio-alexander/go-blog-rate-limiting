package logic

import (
	"fmt"
	"sync"
	"time"

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

type LeakyBucket struct {
	sync.RWMutex
	sync.WaitGroup
	stopper   chan struct{}
	buckets   map[string]queue
	queueSize int
	leakRate  time.Duration
}

func NewLeakyBucket(queueSize int, leakRate time.Duration) *LeakyBucket {
	l := &LeakyBucket{
		buckets:   make(map[string]queue),
		stopper:   make(chan struct{}),
		queueSize: queueSize,
		leakRate:  leakRate,
	}
	return l
}

func (l *LeakyBucket) launchHandler(q queue) {
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

func (l *LeakyBucket) Limit(id string) (bool, <-chan struct{}) {
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
		return true, nil
	}
	fmt.Printf(leakyBucketPrefix+"%s allowed (%d)\n", id, queue.Length())
	return false, queue.GetSignalOut()
}

func (l *LeakyBucket) Stop() {
	l.Lock()
	defer l.Unlock()

	close(l.stopper)
	l.Wait()
	fmt.Println(leakyBucketPrefix + "stopped")
}
