package main

import (
	"log"
	"net/http"
)

func main() {
	s := newServer()
	go s.run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(s, w, r)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
