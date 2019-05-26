package app

import (
	"fmt"
	influx "github.com/influxdata/influxdb1-client/v2"
	"log"
	"time"
)

// DBClient represents a generic interface for DB related operations
type DBClient interface {
	Ping() error
	Save(event EventModel) error
	FetchAll(start int64, end int64) ([]EventModel, error)
}

// InfluxDBClient the implementation of DBClient backed by InfluxDB
type InfluxDBClient struct {
	influxHTTPClient influx.Client
}

// NewInfluxDBClient creates a new instance of InfluxDBClient with embedded influxHTTPClient.
// Returns a value because InfluxDBClient is a small stateless struct.
func NewInfluxDBClient(influxHTTPClient influx.Client) DBClient {
	return InfluxDBClient{
		influxHTTPClient,
	}
}

// Ping pings InfluxDB just to check if connection is OK
func (c InfluxDBClient) Ping() error {
	log.Printf("Ping InfluxDB")
	_, _, err := c.influxHTTPClient.Ping(0)
	return err
}

// Save implements DBClient.Save by saving the specified event into influxDB
func (c InfluxDBClient) Save(event EventModel) error {
	log.Printf("Save event %+v", event)

	bpc := influx.BatchPointsConfig{
		Database: "mydb",
	}
	bps, _ := influx.NewBatchPoints(bpc)

	tags := make(map[string]string)
	tags["event_type"] = event.EventType

	fields := make(map[string]interface{})
	for k, v := range event.Params {
		fields[k] = v
	}

	tm := time.Unix(event.Ts, 0)

	point, err := influx.NewPoint("events", tags, fields, tm)
	if err != nil {
		log.Print(err)
		return err
	}

	bps.AddPoint(point)

	defer c.influxHTTPClient.Close()

	err = c.influxHTTPClient.Write(bps)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

// FetchAll implements DBClient.FetchAll by featching events from influxDB in the specified time range
func (c InfluxDBClient) FetchAll(start int64, end int64) ([]EventModel, error) {
	log.Printf("Fetch all events from now - %ds to now - %ds", start, end)

	cmd := fmt.Sprintf("SELECT * FROM events WHERE time >= NOW() - %ds AND time <= NOW() - %ds", start, end)

	log.Println("Query data with comand", cmd)

	q := influx.Query{
		Command:  cmd,
		Database: "mydb",
	}

	defer c.influxHTTPClient.Close()

	response, err := c.influxHTTPClient.Query(q)
	if err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, response.Error()
	}

	fmt.Printf("%+v \n", response.Results)

	events := []EventModel{}
	for _, result := range response.Results {
		for _, row := range result.Series {
			for _, value := range row.Values {
				tm, err := time.Parse(time.RFC3339, value[0].(string))
				if err != nil {
					log.Print(err)
					return nil, err
				}

				params := make(map[string]interface{})
				for i, v := range value[2:] {
					if v == nil {
						continue
					}
					key := row.Columns[i+2]
					params[key] = v
				}

				e := EventModel{
					Ts:        tm.Unix(),
					EventType: value[1].(string),
					Params:    params,
				}

				events = append(events, e)

				fmt.Println("event:", e)
			}
		}
	}

	return events, nil
}
