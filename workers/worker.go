package workers

import (
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
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

	if c.Context == nil {
		c.Context = context.Background()
	}
	g, ctx := errgroup.WithContext(c.Context)

	//launch workers
	for i := 0; i < numWorkers; i++ {
		w := WorkerID(i)

		//stand a go rountine
		g.Go(func() error {
			for {
				select {
				case tsk := <-tasks:
					if tsk == nil {
						return nil
					}
					err := tsk.Exec(w)
					if err != nil {
						return err
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}
	//wait for all workers done or a worker returns an error
	return g.Wait()
}
