package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			fmt.Println("\nReceived an interrupt, stopping services...\n")
			cleanup()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}

func cleanup() {
	fmt.Println("clean up ...")
}
