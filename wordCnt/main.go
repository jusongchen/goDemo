package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"github.com/jusongchen/goDemo/workers"
	"github.com/pkg/errors"
)

type wordCnt struct {
	source     string
	re         *regexp.Regexp
	numMatches int64
}

//startTimer return a function which calculates elapsed time when called.
func startTimer(name string) func() {
	t := time.Now()
	log.Println(name, "started")
	return func() {
		d := time.Now().Sub(t)
		log.Println(name, "took", d)
	}
}

//implements workers.Task
func (cnt *wordCnt) Exec(w workers.WorkerID) error {
	// stop := startTimer(fmt.Sprintf("worker #%d %s", w, cnt.source))
	defer func() {
		// stop()
		log.Printf("Worker #%d File %s pattern %s found:%d\n", w, cnt.source, cnt.re.String(), cnt.numMatches)
	}()

	f, err := os.Open(cnt.source)
	defer f.Close()

	if err != nil {
		return err
	}
	r := bufio.NewReader(f)

	for {
		loc := cnt.re.FindReaderIndex(r)
		if loc == nil {
			break
		}
		cnt.numMatches++
	}

	return err
}

//taskFunc return a function which makes tasks
func taskFunc(srcFiles []string, re *regexp.Regexp) workers.FactoryFunc {

	var index int
	return func() workers.Task {
		if index == len(srcFiles) { //
			return nil
		}
		name := srcFiles[index]
		index++
		return &wordCnt{source: name, re: re}
	}
}

//FindFiles search directory tree to get files matching regexp pattern
func FindFiles(root string, pattern string) ([]string, error) {

	m := []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.Mode().IsRegular() {
			return nil
		}

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

//DOP degree of parallelism
var (
	DOP         int
	wordPattern string
	re          regexp.Regexp
)

func main() {
	flag.IntVar(&DOP, "DOP", runtime.NumCPU(), "Degree of Parallelism, must be >= 1")
	flag.StringVar(&wordPattern, "e", "", "pattern, must have")

	flag.Usage = func() {
		fmt.Printf("%s by Jusong Chen\n", os.Args[0])
		fmt.Println("Usage:")
		fmt.Printf("   %s [flags] path file \n", os.Args[0])
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	flag.Parse()

	if flag.NArg() != 2 || DOP < 1 || wordPattern == "" {
		flag.Usage()
	}

	re := regexp.MustCompile(".*" + wordPattern + ".*")

	path, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalf("Cannot get absolute path:%s", flag.Arg(0))
	}
	pattern := flag.Arg(1)

	files, err := FindFiles(path, pattern)
	if err != nil {
		log.Fatal(err)
	}

	c := &workers.Context{
		DOP:         DOP,
		FactoryFunc: taskFunc(files, re),
	}

	stop := startTimer(fmt.Sprintf("grep %d files", len(files)))
	defer stop()
	err = workers.Do(c)
	if err != nil {
		log.Fatal(err)
	}

}
