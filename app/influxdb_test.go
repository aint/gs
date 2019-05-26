package app

import (
	"fmt"
	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestFetchByType(t *testing.T) {
	t.Run("Given start, end and type parameters", func(t *testing.T) {
		eventType := "session_start"
		start := int64(1558888555)
		end := int64(1558888888)

		t.Run("Happy case", func(t *testing.T) {
			// given
			response := &influx.Response{Results: []influx.Result{}, Err: ""}

			influxClient := new(MockedInfluxClient)
			influxClient.On("Query", mock.Anything).Return(response, nil)
			influxClient.On("Close").Return(nil)

			dbClient := NewInfluxDBClient(influxClient)

			// when
			events, err := dbClient.FetchByType(eventType, start, end)

			// then
			assert.NoError(t, err)
			assert.Empty(t, events)

			influxClient.AssertExpectations(t)
		})

		t.Run("Fail case", func(t *testing.T) {
			// given
			response := &influx.Response{Results: []influx.Result{}, Err: "Some error on DB read"}

			influxClient := new(MockedInfluxClient)
			influxClient.On("Query", mock.Anything).Return(response, nil)
			influxClient.On("Close").Return(nil)

			dbClient := NewInfluxDBClient(influxClient)

			// when
			events, err := dbClient.FetchByType(eventType, start, end)

			// then
			assert.Error(t, err)
			assert.Empty(t, events)

			influxClient.AssertExpectations(t)
		})
	})
}

func TestSave(t *testing.T) {
	t.Run("Given some events", func(t *testing.T) {
		events := []EventModel{
			EventModel{"type", 123456, map[string]interface{}{"key1": 42}},
			EventModel{"type", 234567, map[string]interface{}{"key2": "42"}},
		}

		t.Run("Happy case", func(t *testing.T) {
			// given
			bps, _ := constructBatchPoints(events)

			influxClient := new(MockedInfluxClient)
			influxClient.On("Write", bps).Return(nil)
			influxClient.On("Close").Return(nil)

			dbClient := NewInfluxDBClient(influxClient)

			// when
			err := dbClient.Save(events)

			// then
			assert.NoError(t, err)

			influxClient.AssertExpectations(t)
		})

		t.Run("Fail case", func(t *testing.T) {
			// given
			influxClient := new(MockedInfluxClient)
			influxClient.On("Write", mock.Anything).Return(fmt.Errorf("Some error on DB write"))
			influxClient.On("Close").Return(nil)

			dbClient := NewInfluxDBClient(influxClient)

			// when
			err := dbClient.Save(events)

			// then
			assert.Error(t, err)

			influxClient.AssertExpectations(t)
		})
	})
}

type MockedInfluxClient struct {
	mock.Mock
}

func (m *MockedInfluxClient) Ping(timeout time.Duration) (time.Duration, string, error) {
	args := m.Called(timeout)
	return args.Get(0).(time.Duration), args.String(1), args.Error(2)
}

func (m *MockedInfluxClient) Write(bp influx.BatchPoints) error {
	args := m.Called(bp)
	return args.Error(0)
}

func (m *MockedInfluxClient) Query(q influx.Query) (*influx.Response, error) {
	args := m.Called(q)
	return args.Get(0).(*influx.Response), args.Error(1)
}

func (m *MockedInfluxClient) QueryAsChunk(q influx.Query) (*influx.ChunkedResponse, error) {
	args := m.Called(q)
	return args.Get(0).(*influx.ChunkedResponse), args.Error(1)
}

func (m *MockedInfluxClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
