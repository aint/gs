package app

import (
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
			mockerDBClient := new(MockedInfluxDBClient)
			mockerDBClient.On("Ping").Return(nil)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Health(mockerDBClient, w, r)
			})

			// when
			handler.ServeHTTP(rr, req)

			// then
			assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

			expected := `{"app_status":"ok","db_status":"ok"}`
			assert.Equal(t, expected, rr.Body.String(), "handler returned unexpected body")

			mockerDBClient.AssertExpectations(t)
		})

		t.Run("Fail case", func(t *testing.T) {
			// given
			mockerDBClient := new(MockedInfluxDBClient)
			mockerDBClient.On("Ping").Return(fmt.Errorf("can't connect to DB"))

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Health(mockerDBClient, w, r)
			})

			// when
			handler.ServeHTTP(rr, req)

			// then
			assert.Equal(t, http.StatusInternalServerError, rr.Code, "handler returned wrong status code")

			expected := `{"error":"Error while pinging DB: 'can't connect to DB'"}`
			assert.Equal(t, expected, rr.Body.String(), "handler returned unexpected body")

			mockerDBClient.AssertExpectations(t)
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
