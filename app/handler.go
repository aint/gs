package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
		errorMessage := fmt.Sprintf("Error while pinging DB: '%s'", err)
		returnError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	healthResponse := map[string]string{"app_status": "ok", "db_status": "ok"}

	returnJSON(w, http.StatusOK, healthResponse)
}

// SaveEvent handles save event request
func SaveEvent(dbClient DBClient, w http.ResponseWriter, r *http.Request) {
	log.Println("Handle save event request")

	event := EventModel{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&event)
	if err != nil {
		log.Println("Error while demarshalling body", err.Error())
		return
	}
	defer r.Body.Close()

	err = dbClient.Save(event)
	if err != nil {
		errorMessage := fmt.Sprintf("Error while saving event to DB: '%s'", err)
		returnError(w, http.StatusInternalServerError, errorMessage)
	}

	w.WriteHeader(http.StatusCreated)
}

// GetEvents handles get events request
func GetEvents(dbClient DBClient, w http.ResponseWriter, r *http.Request) {
	log.Println("Handle get events request")

	start, _ := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
	end, _ := strconv.ParseInt(r.URL.Query().Get("end"), 10, 64)

	events, err := dbClient.FetchAll(start, end)
	if err != nil {
		errorMessage := fmt.Sprintf("Error while fetching events from DB: '%s'", err)
		returnError(w, http.StatusInternalServerError, errorMessage)
	}

	returnJSON(w, http.StatusOK, events)
}

func returnJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error while marshalling payload", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

func returnError(w http.ResponseWriter, status int, message string) {
	log.Printf("Returning error '%s' with status code %d", message, status)
	returnJSON(w, status, map[string]string{"error": message})
}
