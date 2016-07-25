package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Task int
type Worker struct {
	ID      int
	cntTask int
}

type Result struct {
	task       Task
	assignedTo *Worker
	start      time.Time
	elapsed    time.Duration
}

func main() {
	numWorkers := 5
	numTasks := 30

	tasks := make(chan Task)
	//generate tasks
	go func() {
		for i := 0; i < numTasks; i++ {
			tasks <- Task(i)
		}
		close(tasks)
	}()

	results := make(chan Result)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	go func() {
		wg.Wait()
		//when all workers have done their work, close result channel
		close(results)
	}()

	workers := []*Worker{}
	//launch workers
	for i := 0; i < numWorkers; i++ {
		w := &Worker{ID: i}
		workers = append(workers, w)
		go func(w *Worker) {
			defer wg.Done()
			for tsk := range tasks {
				res, err := w.executeTask(tsk)
				if err != nil {
					log.Printf("Worker %v fail when processing %v: %v", w, tsk, err)
					continue
				}
				results <- res
			}
		}(w)
	}

	//handling results
	for r := range results {
		fmt.Printf("Woker #%d completed task %3d in %v\n", r.assignedTo.ID, r.task, r.elapsed)
	}
	//worker summary report
	for _, w := range workers {
		fmt.Printf("Woker #%d:completed %d tasks in total\n", w.ID, w.cntTask)
	}

}

func (w *Worker) executeTask(tsk Task) (Result, error) {

	res := Result{task: tsk, start: time.Now(), assignedTo: w}
	//sleep for random period to simulate task processing
	d := time.Duration(rand.Float64()*2) * time.Second
	time.Sleep(d)
	res.elapsed = time.Since(res.start)
	w.cntTask++
	return res, nil
}
