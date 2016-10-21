package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"github.com/jusongchen/goDemo/workers"
	"github.com/pkg/errors"
)

type gzipCtx struct {
	source string
	target string
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
func (gz *gzipCtx) Exec(w workers.WorkerID) error {
	stop := startTimer(fmt.Sprintf("worker #%d %s", w, gz.source))
	defer stop()

	reader, err := os.Open(gz.source)
	if err != nil {
		return err
	}

	filename := filepath.Base(gz.source)
	writer, err := os.Create(gz.target)
	if err != nil {
		return errors.Wrap(err, "gzip exec")
	}

	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = filename
	defer archiver.Close()

	_, err = io.Copy(archiver, reader)
	return err
}

//taskFunc return a function which makes tasks
func taskFunc(srcFiles []string) workers.FactoryFunc {

	var index int
	return func() workers.Task {
		if index == len(srcFiles) { //
			return nil
		}
		name := srcFiles[index]
		index++
		return &gzipCtx{source: name, target: name + ".gz"}
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
var DOP int

func main() {
	flag.IntVar(&DOP, "DOP", runtime.NumCPU(), "Degree of Parallelism, must be >= 1")

	flag.Usage = func() {
		fmt.Printf("%s by Jusong Chen\n", os.Args[0])
		fmt.Println("Usage:")
		fmt.Printf("   %s [flags] path pattern \n", os.Args[0])
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	flag.Parse()

	if flag.NArg() != 2 || DOP < 1 {
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

	c := &workers.Context{
		DOP:         DOP,
		FactoryFunc: taskFunc(files),
	}

	stop := startTimer(fmt.Sprintf("gzip %d files", len(files)))
	defer stop()
	err = workers.Do(c)
	if err != nil {
		log.Fatal(err)
	}

}
