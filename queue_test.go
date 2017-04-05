package queue

import (
	"sync"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {

	q := New()
	// create 50 producers and 10 consumers
	var wgProd sync.WaitGroup
	producers := 50
	wgProd.Add(producers)
	for i := 0; i < producers; i++ {
		go func(i int) {
			defer wgProd.Done()
			t.Logf("Producer %v: adding: %v\n", i, i)
			q.Add(i)
		}(i)
	}

	var wgCon sync.WaitGroup
	consumers := 10
	wgCon.Add(consumers)
	for i := 0; i < consumers; i++ {
		go func(i int) {
			defer wgCon.Done()
			item, shuttingdown := q.Get()
			if item == "added after shutdown!" {
				t.Errorf("Got an item added after shutdown.")
			}
			if shuttingdown {
				t.Logf("Shuting down...")
				return
			}
			t.Logf("Worker %v: begin processing %v\n", i, item)
			time.Sleep(3 * time.Millisecond)
			t.Logf("Worker %v: done processing %v\n", i, item)
			q.Done(item)
		}(i)
	}

	wgProd.Wait()
	q.ShutDown()
	q.Add("added after shutdown!")
	wgCon.Wait()
}

func TestAddWhileProcessing(t *testing.T) {
	q := New()

	// Start producers
	const producers = 50
	producerWG := sync.WaitGroup{}
	producerWG.Add(producers)
	for i := 0; i < producers; i++ {
		go func(i int) {
			defer producerWG.Done()
			t.Logf("Producer %v: adding: %v\n", i, i)
			q.Add(i)
		}(i)
	}

	// Start consumers
	const consumers = 10
	consumerWG := sync.WaitGroup{}
	consumerWG.Add(consumers)
	// we will add the item we are processing back to the queue
	// this should exercise the dirty map
	for i := 0; i < consumers; i++ {
		go func(i int) {
			defer consumerWG.Done()
			item, done := q.Get()
			if done {
				return
			}
			count := map[interface{}]int{}
			if count[item] < 1 {
				q.Add(item)
				count[item]++
			}
			t.Logf("Worker %v: begin processing %v\n", i, item)
			// time.Sleep(3 * time.Millisecond)
			t.Logf("Worker %v: done processing %v\n", i, item)
			q.Done(item)

		}(i)
	}

	producerWG.Wait()
	q.ShutDown()
	consumerWG.Wait()
}
