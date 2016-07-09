package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Task int

func worker(workerID int, tasksChannel <-chan Task, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		task, ok := <-tasksChannel
		if !ok {
			//no job left on the tasksChannel
			return
		}

		t := time.Now()

		//sleep for random period to simulate task processing
		d := time.Duration(rand.Float64()*2) * time.Second
		time.Sleep(d)

		fmt.Printf("Woker #%d:completed task %d in %s\n", workerID, task, time.Since(t))
	}
}

func pool(wg *sync.WaitGroup, numWorkers, numTasks int) {

	tasksChannel := make(chan Task)

	for i := 0; i < numWorkers; i++ {
		go worker(i, tasksChannel, wg)
	}

	//this goroutine generate jobs
	for i := 0; i < numTasks; i++ {
		tasksChannel <- Task(i)
	}

	close(tasksChannel)
}

func main() {
	var wg sync.WaitGroup

	numWorkers := 8
	numTasks := 50

	wg.Add(numWorkers)

	//launch a goroutine to generate and execute jobs
	go pool(&wg, numWorkers, numTasks)

	//wait until all workers returns
	wg.Wait()
}
