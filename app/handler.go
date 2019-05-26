package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
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

	start, err := validateQueryPeriodParam(r, "start", true)
	if err != nil {
		returnError(w, http.StatusBadRequest, err.Error())
		return
	}

	end, err := validateQueryPeriodParam(r, "end", false)
	if err != nil {
		returnError(w, http.StatusBadRequest, err.Error())
		return
	}

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

func validateQueryPeriodParam(r *http.Request, paramName string, mandatory bool) (int64, error) {
	paramValue := r.URL.Query().Get(paramName)
	log.Printf("Validating query param %s = %s", paramName, paramValue)

	if len(paramValue) == 0 {
		if mandatory {
			return -1, fmt.Errorf("`%s` param should be present", paramName)
		}
		paramValue = "0"
	}
	param, err := parsePeriodToSeconds(paramValue)
	if err != nil {
		return -1, fmt.Errorf("`%s` param is invalid", paramName)
	}

	return param, nil
}

var re = regexp.MustCompile(`(\d+)(m|minute|minutes|h|hour|hours|d|day|days|w|weeks|weeks)`)

func parsePeriodToSeconds(period string) (int64, error) {
	if re.MatchString(period) {
		groups := re.FindStringSubmatch(period)
		// no need to handle error as it already matched by regexp
		amount, _ := strconv.Atoi(groups[1])

		switch unit := groups[2]; unit {
		case "m":
			log.Println("minutes")
			return int64(amount * 60), nil
		case "h":
			log.Println("hours")
			return int64(amount * 3600), nil
		case "d":
			log.Println("days")
			return int64(amount * 86400), nil
		case "w":
			log.Println("days")
			return int64(amount * 604800), nil
		default:
			panic("Specified period can't be processed")
		}

	}

	return strconv.ParseInt(period, 10, 64)
}
