package app

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHealth(t *testing.T) {
	t.Run("Given a request to health endpoint", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/health", nil)
		assert.NoError(t, err)

		t.Run("Happy case", func(t *testing.T) {
			// given
			mockedDBClient := new(MockedInfluxDBClient)
			mockedDBClient.On("Ping").Return(nil)

			rr := httptest.NewRecorder()

			// when
			Health(mockedDBClient, rr, req)

			// then
			assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

			expected := `{"app_status":"ok","db_status":"ok"}`
			assert.Equal(t, expected, rr.Body.String(), "handler returned unexpected body")

			mockedDBClient.AssertExpectations(t)
		})

		t.Run("Fail case", func(t *testing.T) {
			// given
			mockedDBClient := new(MockedInfluxDBClient)
			mockedDBClient.On("Ping").Return(fmt.Errorf("can't connect to DB"))

			rr := httptest.NewRecorder()

			// when
			Health(mockedDBClient, rr, req)

			// then
			assert.Equal(t, http.StatusInternalServerError, rr.Code, "handler returned wrong status code")

			expected := `{"error":"Error while pinging DB: 'can't connect to DB'"}`
			assert.Equal(t, expected, rr.Body.String(), "handler returned unexpected body")

			mockedDBClient.AssertExpectations(t)
		})
	})
}

func TestSaveEvents(t *testing.T) {
	t.Run("Given a request and body for save events endpoint", func(t *testing.T) {
		body := `[{
			"event_type": "link_clicked",
			"ts": 1558892660,
			"params": { "url": "localhost:5000/app" }
		},
		{
			"event_type": "link_clicked",
			"ts": 1558892797,
			"params": { "url": "localhost:6000/app" }
		}]`
		req, err := http.NewRequest(http.MethodPost, "/events", bytes.NewBuffer([]byte(body)))
		assert.NoError(t, err)

		t.Run("Happy case", func(t *testing.T) {
			// given
			mockedDBClient := new(MockedInfluxDBClient)
			mockedDBClient.On("Save", mock.Anything).Return(nil)

			rr := httptest.NewRecorder()

			// when
			SaveEvents(mockedDBClient, rr, req)

			// then
			assert.Equal(t, http.StatusCreated, rr.Code, "handler returned wrong status code")
			assert.Equal(t, "", rr.Body.String(), "handler returned unexpected body")

			mockedDBClient.AssertExpectations(t)
		})
	})
}

func TestGetEvents(t *testing.T) {
	t.Run("Given a request with params to get events endpoint", func(t *testing.T) {
		eventType := "session_created"
		start := 10
		end := 5
		url := fmt.Sprintf("/events/relative?type=%s&start=%dm&end=%dm", eventType, start, end)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		assert.NoError(t, err)

		t.Run("Happy case", func(t *testing.T) {
			// given
			events := []EventModel{
				EventModel{eventType, int64(1558892660), map[string]interface{}{"key1": "val1"}},
				EventModel{eventType, int64(1558892797), map[string]interface{}{"key2": "val2"}},
			}
			mockedDBClient := new(MockedInfluxDBClient)
			mockedDBClient.On("FetchByType", eventType, int64(60 * start), int64(60 * end)).Return(events, nil)

			rr := httptest.NewRecorder()

			// when
			GetEvents(mockedDBClient, rr, req)

			// then
			assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

			expected := "[{\"event_type\":\"session_created\",\"ts\":1558892660,\"params\":{\"key1\":\"val1\"}},{\"event_type\":\"session_created\",\"ts\":1558892797,\"params\":{\"key2\":\"val2\"}}]"
			assert.Equal(t, expected, rr.Body.String(), "handler returned unexpected body")

			mockedDBClient.AssertExpectations(t)
		})
	})
}

type MockedInfluxDBClient struct {
	mock.Mock
}

func (m *MockedInfluxDBClient) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockedInfluxDBClient) Save(event []EventModel) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockedInfluxDBClient) FetchAll(start int64, end int64) ([]EventModel, error) {
	args := m.Called(start, end)
	return args.Get(0).([]EventModel), args.Error(1)
}

func (m *MockedInfluxDBClient) FetchByType(eventType string, start int64, end int64) ([]EventModel, error) {
	args := m.Called(eventType, start, end)
	return args.Get(0).([]EventModel), args.Error(1)
}
