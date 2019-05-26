package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/health", Health).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Health the health endpoint provides basic application health information
func Health(w http.ResponseWriter, r *http.Request) {
	log.Println("Health endpoint")

	healthStruct := struct {
		AppStatus string `json:"app_status"`
	}{
		"ok",
	}

	response, err := json.Marshal(healthStruct)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Println("Error while marshaling health struct", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}
