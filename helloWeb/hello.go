package main

import (
	"fmt"
	"log"
	"net/http"
)

func greetingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi %s, how are you doing today?", r.URL.Path[1:])
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "bad request:%s!", r.URL.Path[:])
}

func main() {
	http.HandleFunc("/hello/", greetingHandler)
	http.HandleFunc("/", defaultHandler)
	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
