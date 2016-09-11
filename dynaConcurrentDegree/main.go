package main

import (
	"fmt"
	"sync"
	"time"
)

type Task int

type Result struct {
	task    Task
	start   time.Time
	elapsed time.Duration
}

var (
	semaphore chan struct{}
	results   chan Result

	//MDOP is the max degree of parallism, unexported
	MDOP     = 3
	numTasks = 3020
)

func main() {

	semaphore = make(chan struct{}, MDOP)
	tasks := make(chan Task)
	results = make(chan Result)

	go tuner()

	//generate tasks
	go func() {
		for i := 0; i < numTasks; i++ {
			tasks <- Task(i)
		}
		close(tasks)
	}()

	go func() {
		for r := range results {
			fmt.Printf("task %3d completed in %v\n", r.task, r.elapsed)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(numTasks)

	for tsk := range tasks {
		semaphore <- struct{}{} //this will block when channel is full

		go func(tsk Task) {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			res := Result{task: tsk, start: time.Now()}
			//sleep for random period to simulate task processing
			// d := time.Duration(rand.Int31n(10)) * time.Millisecond
			time.Sleep(time.Second)
			res.elapsed = time.Since(res.start)
			results <- res
		}(tsk)

	}
	wg.Wait()
	close(results)
	close(semaphore)
}

func tuner() {
	c := time.Tick(time.Second * 5)
	for now := range c {
		AdjAmt := now.Second()/10 - 2
		fmt.Printf("\nNow at AdjAmt %d\n==============================\n", AdjAmt)
		adjust(AdjAmt)
	}
}

func adjust(AdjAmt int) {
	for {
		switch {
		case AdjAmt == 0:
			return
		case AdjAmt > 0:
			<-semaphore
			AdjAmt--
		case AdjAmt < 0:
			semaphore <- struct{}{}
			AdjAmt++
		}
	}
}
