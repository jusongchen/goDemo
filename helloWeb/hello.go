package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	http.HandleFunc("/hello/", greetingHandler)
	http.HandleFunc("/", defaultHandler)

	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}

func greetingHandler(w http.ResponseWriter, r *http.Request) {
	//split url path to two strings using / as delimitor
	name := strings.SplitN(r.URL.Path[1:], "/", 2)[1]
	fmt.Fprintf(w, "Hi %s,how are you today?", name)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("bad request:%s", r.URL.Path[:]), http.StatusBadRequest)
}
