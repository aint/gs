package main

import (
	"github.com/aint/gs/app"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/health", app.Health).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
