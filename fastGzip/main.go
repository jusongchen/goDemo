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

	"github.com/jusongchen/goDemo/workers"
	"github.com/pkg/errors"
)

//DOP degree of parallelism
var DOP int

type gzipCtx struct {
	source string
	target string
}

//Task implements exec method
func (gz *gzipCtx) Exec() error {
	// log.Printf("processing:%s", gz.source)
	// return nil

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

//Task implements exec method
func (gz *gzipCtx) PreExec(exec workers.TaskExec) {
	// log.Printf("Task begin\t\t worker #%d \t:%s\n", exec.WorkerID, gz.source)
	log.Printf("Task begin\t\t Woker #%d \t %v \t\t%s\n", exec.WorkerID, exec.Elapsed, gz.source)
}

func (gz *gzipCtx) PostExec(exec workers.TaskExec) {
	log.Printf("Task end\t\t Woker #%d \t %v \t\t%s\n", exec.WorkerID, exec.Elapsed, gz.target)
}

type fileList []string

var fileIndex int

func (l fileList) Make() workers.Task {
	if fileIndex == len(l) {
		return nil
	}
	name := l[fileIndex]
	fileIndex++
	return &gzipCtx{source: name, target: name + ".gz"}
}

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

	l, err := FindFiles(path, pattern)
	if err != nil {
		log.Fatal(err)
	}
	// log.Printf("files Found:%v\n", l)
	// c := workers.Context{}

	err = workers.Do(DOP, fileList(l))
	if err != nil {
		log.Fatal(err)
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
