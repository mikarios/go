package main

import (
	"fmt"
	"log"
	"net/http"
)

func httpServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func main() {
	http.HandleFunc("/", httpServer)
	log.Fatal(http.ListenAndServe(":80", nil))
}
