// Package workqueue provides a simple queue that supports the following
// features:
//  * Fair: items processed in the order in which they are added.
//  * Stingy: a single item will not be processed multiple times concurrently,
//      and if an item is added multiple times before it can be processed, it
//      will only be processed once.
//  * Multiple consumers and producers. In particular, it is allowed for an
//      item to be reenqueued while it is being processed.
//  * Shutdown notifications.
package queue

import "sync"

type Interface interface {
	Add(item interface{})
	Len() int
	Get() (item interface{}, shutdown bool)
	Done(item interface{})
	ShutDown()
	IsShuttingDown() bool
}

type set map[interface{}]struct{}

type Queue struct {
	queue        []interface{}
	dirty        set
	processing   set
	condition    *sync.Cond
	shuttingdown bool
}

func New() *Queue {
	return &Queue{
		dirty:      set{},
		processing: set{},
		condition:  sync.NewCond(&sync.Mutex{}),
	}
}

func (q *Queue) Add(item interface{}) {

	q.condition.L.Lock()
	defer q.condition.L.Unlock()
	if q.shuttingdown {
		return
	}
	if _, exists := q.dirty[item]; exists {
		return
	}
	q.dirty[item] = struct{}{}
	if _, exists := q.processing[item]; exists {
		// we'll add it from dirty when Done() is called
		return
	}
	q.queue = append(q.queue, item)
	q.condition.Signal()
}

func (q *Queue) Get() (interface{}, bool) {
	q.condition.L.Lock()
	defer q.condition.L.Unlock()
	for len(q.queue) == 0 && !q.shuttingdown {
		// we wait if queue is empty
		// if we're shuttingdown, then we stop waiting and let the queue
		// be drained (we don't allow new items to be added)
		q.condition.Wait()
	}
	if len(q.queue) == 0 {
		// we must be shuttingdown, if queue is empty, notify workers
		return nil, true
	}
	item := q.queue[0]
	q.queue = q.queue[1:]
	q.processing[item] = struct{}{} // add to processing
	delete(q.dirty, item)           // remove from dirty
	return item, false
}

func (q *Queue) Done(item interface{}) {
	q.condition.L.Lock()
	defer q.condition.L.Unlock()
	delete(q.processing, item)
	if _, exists := q.dirty[item]; !exists {
		return
	}
	delete(q.dirty, item)
	q.queue = append(q.queue, item)
	q.condition.Signal()
}

func (q *Queue) ShutDown() {
	q.condition.L.Lock()
	defer q.condition.L.Unlock()

	q.shuttingdown = true
	q.condition.Broadcast()
}

func (q *Queue) IsShuttingDown() bool {
	q.condition.L.Lock()
	defer q.condition.L.Unlock()
	return q.shuttingdown
}

func (q *Queue) Len() int {
	q.condition.L.Lock()
	defer q.condition.L.Unlock()
	return len(q.queue)
}
