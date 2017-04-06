package queue

import (
	"fmt"
	"sync"

	utilruntime "k8s.io/kubernetes/pkg/util/runtime"
)

type WorkFunc func(piece interface{})

// ParallelizerSimple parallelizes the processing of "pieces" by the WorkFunc
// It's simple in that it assumes the number of pieces are known before hand
// and will run untill all of them have been processed.
func ParallelizerSimple(workers int, pieces []interface{}, wf WorkFunc) {
	// we write all the pieces into a channel that is exactly the length
	// of the slice containing the pieces
	toProcessChan := make(chan interface{}, len(pieces))
	for _, p := range pieces {
		toProcessChan <- p
	}
	close(toProcessChan)

	var wg sync.WaitGroup
	wg.Add(workers)
	// we create the desired number of workers and have each worker
	// read from the above chan the piece and call the desired WorkFunc
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			defer utilruntime.HandleCrash()
			// Read from channel until it is closed
			for p := range toProcessChan {
				wf(p)
			}
		}()
	}

	wg.Wait()
}

// ParallelizerWithShutDown parallelizes the processing of "pieces" by the
// WorkFunc. It assumes the number of pieces are known before hand
// and will run untill all of them have been processed or it receives a
// shutdown request on the shutdown chan.
func ParallelizerWithShutDown(shutdown <-chan struct{}, workers int, pieces []interface{}, wf WorkFunc) {
	// we write all the pieces into a channel that is exactly the length
	// of the slice containing the pieces
	toProcessChan := make(chan interface{}, len(pieces))
	for _, p := range pieces {
		toProcessChan <- p
	}
	close(toProcessChan)

	var wg sync.WaitGroup
	wg.Add(workers)
	// we create the desired number of workers and have each worker
	// read from the above chan the piece and call the desired WorkFunc
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			defer utilruntime.HandleCrash()
			// Read from channel until it is closed
			for p := range toProcessChan {
				select {
				case <-shutdown:
					fmt.Printf("Worker %v: Shutdown recieved\n", i)
					return
				default:
					wf(p)
				}

			}
		}(i)
	}
	wg.Wait()
}

// ParallelizerWithInputChan parallelizes the processing of "pieces" by the
// WorkFunc. It continues to read from the peices from the input chan, until
// the chan is closed or shutdown is called.
func ParallelizerWithInputChan(shutdown <-chan struct{}, workers int, inChan <-chan interface{}, wf WorkFunc) {

	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			defer utilruntime.HandleCrash()
			for { //p := range inChan {
				select {
				case <-shutdown:
					fmt.Printf("Worker %v: Shutdown recieved\n", i)
					return
				case p := <-inChan:
					wf(p)
				}
			}
		}(i)

	}
	wg.Wait()
}
