package app

import (
	"encoding/json"
	"log"
	"net/http"
)

// EventModel represents event JSON
type EventModel struct {
	EventType string                 `json:"event_type"`
	Ts        int64                  `json:"ts"`
	Params    map[string]interface{} `json:"params"`
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