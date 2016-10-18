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

//DOP degree of parallelism
var DOP int

type gzipCtx struct {
	source string
	target string
}

func (ctrl *gzipCtx) String() string {
	return "gzip  " + ctrl.source
}

//Task implements exec method
func (ctrl *gzipCtx) Exec() error {
	// log.Printf("processing:%s", ctrl.source)
	// return nil

	reader, err := os.Open(ctrl.source)
	if err != nil {
		return err
	}

	filename := filepath.Base(ctrl.source)
	writer, err := os.Create(ctrl.target)
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

type fileList []string

var i int

func (l fileList) Make() workers.Task {
	if i == len(l) {
		return nil
	}
	name := l[i]
	i++
	return &gzipCtx{source: name, target: name + ".gz"}
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

	l, err := FindFiles(path, pattern)
	if err != nil {
		log.Fatal(err)
	}
	start := time.Now()
	log.Printf("\nStart gzip at %v\n", start)
	err = workers.Do(DOP, fileList(l))
	log.Printf("Total Elapsed Time:%v\n", time.Since(start))
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
