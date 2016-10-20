package workers

import (
	"sync"

	"golang.org/x/net/context"
)

type (
	//WorkerID is exported and identifies a goroutine
	WorkerID int

	//Task is the interface which executable tasks must implement
	Task interface {
		Exec(WorkerID) error
	}

	//FactoryFunc is the function to be invoked to make instances of Task
	FactoryFunc func() Task

	// Context specifies controls of concurrent task executions
	Context struct {
		context.Context
		DOP int
		FactoryFunc
	}
)

//Do execute tasks in parallel
func Do(c *Context) error {
	numWorkers := c.DOP

	tasks := make(chan Task)
	//generate tasks
	go func() {
		for {
			task := c.FactoryFunc()
			if task == nil { //no more tasks
				close(tasks)
				return
			}
			tasks <- task
		}
	}()

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	//launch workers
	for i := 0; i < numWorkers; i++ {
		w := WorkerID(i)

		go func(w WorkerID) {
			defer wg.Done()
			for tsk := range tasks {
				tsk.Exec(w)
			}
		}(w)
	}

	//wait for all workers done
	wg.Wait()

	return nil
}
