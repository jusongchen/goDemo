package workers

import (
	"log"
	"sync"
	"time"
)

//DOP degree of parallelism
var DOP int

//Task interface
type Task interface {
	Exec() error
	String() string
}

//Worker keep a tract of number of tasks executed
type Worker struct {
	WorkerID        int
	cntTask         int
	workingDuration time.Duration
}

//TaskExec task execution stat
type TaskExec struct {
	Task
	*Worker
	start   time.Time
	Elapsed time.Duration
}

//Factory make tasks
type Factory interface {
	Make() Task
}

//Do execute tasks in parallel
func Do(DOP int, f Factory) ([]TaskExec, error) {
	numWorkers := DOP

	tasks := make(chan Task)
	//generate tasks
	go func() {
		for {
			task := f.Make()
			if task == nil { //no more tasks
				close(tasks)
				return
			}
			tasks <- task
		}
	}()

	taskExecs := make(chan TaskExec)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	go func() {
		wg.Wait()
		//when all workers have done their work, close TaskExec channel
		close(taskExecs)
	}()

	workers := []*Worker{}
	//launch workers
	for i := 0; i < numWorkers; i++ {
		w := &Worker{WorkerID: i}
		workers = append(workers, w)
		go func(w *Worker) {
			defer wg.Done()
			for tsk := range tasks {

				taskExec := TaskExec{Task: tsk, Worker: w, start: time.Now()}
				err := tsk.Exec()
				taskExec.Elapsed = time.Since(taskExec.start)

				//Instrument worker executions
				taskExec.cntTask++
				taskExec.workingDuration += taskExec.Elapsed

				if err != nil {
					log.Printf("Worker %v fail when processing %v: %v", w, tsk, err)
					continue
				}
				taskExecs <- taskExec
			}
		}(w)
	}

	exec := []TaskExec{}
	for e := range taskExecs {
		exec = append(exec, e)
	}
	return exec, nil
}
