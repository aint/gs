package app

import (
	"strconv"
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
func Health(dbClient DBClient, w http.ResponseWriter, r *http.Request) {
	log.Println("Health endpoint")

	err := dbClient.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Println("Error while pinging DB", err.Error())
		return
	}

	healthStruct := struct {
		AppStatus string `json:"app_status"`
		DBStatus  string `json:"db_status"`
	}{
		"ok", "ok",
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

// SaveEvent handles save event request
func SaveEvent(dbClient DBClient, w http.ResponseWriter, r *http.Request) {
	log.Println("Handle save event request")

	event := EventModel{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&event); err != nil {
		log.Println("Error while marshaling event struct", err.Error())
		return
	}
	defer r.Body.Close()

	err := dbClient.Save(event)
	if err != nil {
		log.Println("Error while saving event into DB", err)
		//return error json
	}
}

// GetEvents handles get events request
func GetEvents(dbClient DBClient, w http.ResponseWriter, r *http.Request) {
	log.Println("Handle get events request")

	start, _ := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
	end, _ := strconv.ParseInt(r.URL.Query().Get("end"), 10, 64)

	events, err := dbClient.FetchAll(start, end)
	if err != nil {
		log.Println("Error while saving event into DB", err)
		//return error json
	}

	response, err := json.Marshal(events)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Println("Error while marshaling array of events", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}
