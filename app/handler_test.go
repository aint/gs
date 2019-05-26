package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestHealth(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	mockerDBClient := new(MockedInfluxDBClient)
	mockerDBClient.On("Ping").Return(nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Health(mockerDBClient, w, r)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"app_status":"ok","db_status":"ok"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

type MockedInfluxDBClient struct {
	mock.Mock
}

func (m *MockedInfluxDBClient) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockedInfluxDBClient) Save(event EventModel) error {
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
