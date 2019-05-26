package app

import (
	"fmt"
	influx "github.com/influxdata/influxdb1-client/v2"
	"log"
	"time"
)

const (
	databaseName    = "mydb"
	eventsTableName = "events"
)

// DBClient represents a generic interface for DB related operations
type DBClient interface {
	Ping() error
	Save(event []EventModel) error
	FetchAll(start int64, end int64) ([]EventModel, error)
	FetchByType(eventType string, start int64, end int64) ([]EventModel, error)
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

// Save implements DBClient.Save by batch saving the specified events into influxDB
func (c InfluxDBClient) Save(events []EventModel) error {
	log.Printf("Save events: %+v", events)

	bps, err := constructBatchPoints(events)
	if err != nil {
		log.Print("Error while creating batch points for InfluxDB", err)
		return err
	}

	defer c.influxHTTPClient.Close()

	err = c.influxHTTPClient.Write(bps)
	if err != nil {
		log.Print("Error while writting data to InfluxDB", err)
		return err
	}

	return nil
}

func constructBatchPoints(events []EventModel) (influx.BatchPoints, error) {
	bpc := influx.BatchPointsConfig{
		Database: databaseName,
	}

	bps, err := influx.NewBatchPoints(bpc)
	if err != nil {
		return nil, err
	}

	for _, e := range events {
		tags := map[string]string{"event_type": e.EventType}

		fields := make(map[string]interface{})
		for k, v := range e.Params {
			fields[k] = v
		}

		tm := time.Unix(e.Ts, 0)

		point, err := influx.NewPoint(eventsTableName, tags, fields, tm)
		if err != nil {
			return nil, err
		}

		bps.AddPoint(point)
	}

	return bps, nil
}

// FetchAll implements DBClient.FetchAll by featching events from influxDB in the specified time range
func (c InfluxDBClient) FetchAll(start int64, end int64) ([]EventModel, error) {
	log.Printf("Fetch all events from now - %ds to now - %ds", start, end)

	cmd := fmt.Sprintf(`SELECT * FROM %s
						WHERE time >= NOW() - %ds AND time <= NOW() - %ds`, eventsTableName, start, end)

	log.Println("Query data with command", cmd)

	response, err := c.queryDB(cmd)
	if err != nil {
		return nil, err
	}

	return c.parseResponse(response)
}

// FetchByType implements DBClient.FetchByType by featching events by type from influxDB in the specified time range
func (c InfluxDBClient) FetchByType(eventType string, start int64, end int64) ([]EventModel, error) {
	log.Printf("Fetch events by type %s and from now - %ds to now - %ds", eventType, start, end)

	cmd := fmt.Sprintf(`SELECT * FROM %s
						WHERE event_type='%s'
								AND time >= NOW() - %ds
								AND time <= NOW() - %ds`, eventsTableName, eventType, start, end)

	log.Println("Query data with command", cmd)

	response, err := c.queryDB(cmd)
	if err != nil {
		return nil, err
	}

	return c.parseResponse(response)
}

// queryDB convenience function to query the database
func (c InfluxDBClient) queryDB(cmd string) (*influx.Response, error) {
	defer c.influxHTTPClient.Close()

	q := influx.Query{
		Command:  cmd,
		Database: databaseName,
	}

	response, err := c.influxHTTPClient.Query(q)
	if err != nil {
		log.Print("Error while querying InfluxDB", err)
		return nil, err
	}
	if response.Error() != nil {
		log.Print("Error in response from InfluxDB", err)
		return nil, response.Error()
	}

	return response, nil
}

func (c InfluxDBClient) parseResponse(response *influx.Response) ([]EventModel, error) {
	events := []EventModel{}
	for _, result := range response.Results {
		for _, row := range result.Series {
			for _, value := range row.Values {
				tm, err := time.Parse(time.RFC3339, value[0].(string))
				if err != nil {
					log.Print("Error while parsing time from InfluxDB")
					return events, err
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
			}
		}
	}

	return events, nil
}
