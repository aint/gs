package main

import (
	"github.com/aint/gs/app"
	"github.com/gorilla/mux"
	influx "github.com/influxdata/influxdb1-client/v2"
	"log"
	"net/http"
	"os"
)

// App represents the application abstraction with underlying DB client and HTTP server
type App struct {
	router   *mux.Router
	dbClient app.DBClient
}

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "http://localhost:8086"
	}

	influxClient := newInfluxHTTPClient(dbURL)
	dbClient := app.NewInfluxDBClient(influxClient)
	router := mux.NewRouter()

	app := New(router, dbClient)
	app.setRouters()
	app.startHTTPServer(":8080")
}

// New returns a new instance of App by injecting dependencies
func New(router *mux.Router, dbClient app.DBClient) App {
	return App{
		router, dbClient,
	}
}

func (a App) startHTTPServer(port string) {
	log.Println("App is starting ...")
	log.Fatal(http.ListenAndServe(port, a.router))
}

func (a App) setRouters() {
	a.router.HandleFunc("/events/relative", a.getEventsEndpoint).Methods(http.MethodGet)
	a.router.HandleFunc("/events", a.postEventsEndpoint).Methods(http.MethodPost)
	a.router.HandleFunc("/health", a.getHealthEndpoint).Methods(http.MethodGet)
}

func (a App) getHealthEndpoint(w http.ResponseWriter, r *http.Request) {
	app.Health(a.dbClient, w, r)
}

func (a App) getEventsEndpoint(w http.ResponseWriter, r *http.Request) {
	app.GetEvents(a.dbClient, w, r)
}

func (a App) postEventsEndpoint(w http.ResponseWriter, r *http.Request) {
	app.SaveEvents(a.dbClient, w, r)
}

func newInfluxHTTPClient(URL string) influx.Client {
	httpClient, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr: URL,
	})
	if err != nil {
		log.Fatal("Error creating InfluxDB HTTP client: ", err.Error())
	}

	return httpClient
}
