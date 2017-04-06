package queue

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func printMe(p interface{}) {
	fmt.Printf("%v\n", p)
	time.Sleep(time.Millisecond)
}

func printMeJittered(p interface{}) {
	wait := time.Millisecond + time.Duration(rand.Float64()*float64(time.Millisecond))
	fmt.Printf("%v (%v)\n", p, wait)
	time.Sleep(wait)
}

func TestBasic_ParallelizerSimple(t *testing.T) {

	pieces := []int{1, 2, 3, 4, 5}
	piecesGoIf := make([]interface{}, len(pieces))
	for i, _ := range pieces {
		piecesGoIf[i] = pieces[i]
	}
	const workers = 2
	ParallelizerSimple(workers, piecesGoIf, printMe)

}

func TestBasic_ParallelizerWithShutDown(t *testing.T) {

	pieces := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	piecesGoIf := make([]interface{}, len(pieces))
	for i, _ := range pieces {
		piecesGoIf[i] = pieces[i]
	}
	const workers = 2
	downChan := make(chan struct{})
	go ParallelizerWithShutDown(downChan, workers, piecesGoIf, printMe)

	time.Sleep(3 * time.Millisecond)
	// downChan <- struct{}{}
	close(downChan)
	time.Sleep(2 * time.Millisecond)
}

func gen(shutdown <-chan struct{}) <-chan interface{} {
	out := make(chan interface{}, 10)
	i := 0
	go func(i int) {
		for {
			select {
			case <-shutdown:
				fmt.Printf("gen() %v \n", i)
				return
			default:
				out <- i
				i++
				wait := time.Millisecond
				time.Sleep(wait)
			}
		}
	}(i)
	return out
}

func TestBasic_ParallelizerWithInputChan(t *testing.T) {

	downChan := make(chan struct{})
	inChan := gen(downChan)
	const workers = 2
	go ParallelizerWithInputChan(downChan, workers, inChan, printMeJittered)

	time.Sleep(9 * time.Millisecond)
	// downChan <- struct{}{}
	close(downChan)
	time.Sleep(2 * time.Millisecond)
}
