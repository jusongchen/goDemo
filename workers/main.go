package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func worker(workerID int, tasksCh <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		task, ok := <-tasksCh
		if !ok {
			return
		}
		//		d := time.Duration(task) * time.Millisecond
		d := time.Duration(rand.Float64()*2) * time.Second
		t := time.Now()
		time.Sleep(d)
		fmt.Printf("Woker #%d:completed task %d in %s\n", workerID, task, time.Since(t))
	}
}

func pool(wg *sync.WaitGroup, workers, tasks int) {
	tasksCh := make(chan int)

	for i := 0; i < workers; i++ {
		go worker(i, tasksCh, wg)
	}

	for i := 0; i < tasks; i++ {
		tasksCh <- i
	}

	close(tasksCh)
}

func main() {
	var wg sync.WaitGroup
	numWorkers := 8
	numTasks := 50

	wg.Add(numWorkers)
	go pool(&wg, numWorkers, numTasks)
	wg.Wait()
}
