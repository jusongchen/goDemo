package workers

import (
	"sync"

	"golang.org/x/net/context"
)

type (
	//WorkerID exported
	WorkerID int

	//Task interface
	Task interface {
		Exec(WorkerID) error
	}

	//Factory make tasks to be executed by workers in parallel
	Factory interface {
		Make() Task
	}

	// Context controls tasks execution
	Context struct {
		context.Context
		DOP int
		Factory
	}
)

//Do execute tasks in parallel
func (c *Context) Do() error {
	numWorkers := c.DOP

	tasks := make(chan Task)
	//generate tasks
	go func() {
		for {
			task := c.Factory.Make()
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
