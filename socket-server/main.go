package main

import (
	"encoding/json"
	"fmt"
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

func RespJSON(w http.ResponseWriter, r bool, e error) {
	var code int  //http status code
	var es string //error string
	if r == true {
		code = http.StatusAccepted
		es = ""
	} else {
		code = http.StatusUnauthorized
		es = e.Error()
	}

	re := Resp{es}
	result, err := json.Marshal(re)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(result)
}
