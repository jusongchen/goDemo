package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
)

//DOP degree of parallelism
var DOP int

//Task interface
type Task interface {
	exec() error
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
	elapsed time.Duration
}

//MD5Sum MD5 hash of a file
type MD5Sum struct {
	filename string
	MD5      [md5.Size]byte
}

//Task implements exec method
func (task *MD5Sum) exec() error {

	data, err := ioutil.ReadFile(task.filename)
	if err != nil {
		return errors.Wrap(err, "MD5Sum ReadFile")
	}
	task.MD5 = md5.Sum(data)
	return nil
}

func (task *MD5Sum) String() string {

	return task.filename + " MD5:" + fmt.Sprintf("%x", task.MD5)
}

func main() {
	flag.IntVar(&DOP, "DOP", runtime.NumCPU(), "Degree of Parallelism")

	flag.Usage = func() {
		fmt.Printf("%s by Jusong Chen\n", os.Args[0])
		fmt.Println("Usage:")
		fmt.Printf("   %s [flags] path pattern \n", os.Args[0])
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
	}
	path, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalf("Cannot get absolute path:%s", flag.Arg(0))
	}
	pattern := flag.Arg(1)

	files, err := FindFiles(path, pattern)
	if err != nil {
		log.Fatal(err)
	}
	if err = MD5files(DOP, files); err != nil {
		log.Fatal(err)
	}
}

//MD5files calculate MD5 of files in parallel
func MD5files(DOP int, files []string) error {
	numWorkers := DOP

	tasks := make(chan MD5Sum)
	//generate tasks
	go func() {
		for i := range files {
			tasks <- MD5Sum{filename: files[i]}
		}
		close(tasks)
	}()

	TaskExecs := make(chan TaskExec)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	go func() {
		wg.Wait()
		//when all workers have done their work, close TaskExec channel
		close(TaskExecs)
	}()

	workers := []*Worker{}
	//launch workers
	for i := 0; i < numWorkers; i++ {
		w := &Worker{WorkerID: i}
		workers = append(workers, w)
		go func(w *Worker) {
			defer wg.Done()
			for tsk := range tasks {

				taskExec := TaskExec{Task: &tsk, Worker: w, start: time.Now()}
				err := tsk.exec()
				taskExec.elapsed = time.Since(taskExec.start)
				taskExec.cntTask++
				taskExec.workingDuration += taskExec.elapsed

				if err != nil {
					log.Printf("Worker %v fail when processing %v: %v", w, tsk, err)
					continue
				}
				TaskExecs <- taskExec
			}
		}(w)
	}

	//handling TaskExecs
	for r := range TaskExecs {
		fmt.Printf("Woker #%d took %v to complete task: %v\n", r.WorkerID, r.elapsed, r.Task)
	}
	//worker summary report
	for _, w := range workers {
		fmt.Printf("Woker #%d completed %d tasks in %v\n", w.WorkerID, w.cntTask, w.workingDuration)
	}
	return nil
}

//FindFiles search directory tree to get files matching regexp pattern
func FindFiles(root string, pattern string) ([]string, error) {

	m := []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.Mode().IsRegular() {
			return nil
		}

		// if err == os.ErrPermission {
		// 	return nil
		// }
		if err != nil {
			return errors.Wrap(err, "filepath Walk")
		}

		matched, err := regexp.MatchString(pattern, info.Name())
		if err != nil {
			return errors.Wrap(err, "FindFiles regexp")
		}
		// matched := true
		if matched {
			// fmt.Println("Find file:", path)
			m = append(m, path)
		}
		return nil
	})
	// fmt.Printf("Files:%v", m)
	return m, err
}
