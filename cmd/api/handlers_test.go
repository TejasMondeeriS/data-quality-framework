package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"xcaliber/data-quality-metrics-framework/internal/database"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) AddNewQueries(query database.Query) error {
	args := m.Called(query)
	return args.Error(0)
}

func (m *MockDB) FetchQUeryString(queryID uuid.UUID) (string, error) {
	args := m.Called(queryID)
	return args.String(0), args.Error(1)
}

func (m *MockDB) FetchAllQUeries() ([]database.Query, error) {
	args := m.Called()
	return args.Get(0).([]database.Query), args.Error(1)
}

func TestAddQuery(t *testing.T) {
	tests := []struct {
		name           string
		input          AddQueryRequest
		setupMock      func(*MockDB)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Valid Query Creation",
			input: AddQueryRequest{
				Name:              "Test Query",
				Description:       "Test Description",
				Query:             "SELECT * FROM test",
				DefaultParameters: json.RawMessage(`{"param":"value"}`),
				DataProductID:     uuid.New(),
			},
			setupMock: func(m *MockDB) {
				m.On("AddNewQueries", mock.AnythingOfType("database.Query")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "Database Error",
			input: AddQueryRequest{
				Name:              "Test Query",
				Description:       "Test Description",
				Query:             "SELECT * FROM test",
				DefaultParameters: json.RawMessage(`{"param":"value"}`),
				DataProductID:     uuid.New(),
			},
			setupMock: func(m *MockDB) {
				m.On("AddNewQueries", mock.AnythingOfType("database.Query")).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.setupMock(mockDB)

			newLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
			app := &application{db: mockDB, logger: newLogger}

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			app.AddQuery(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestValidateAddQueryRequestParameters(t *testing.T) {
	tests := []struct {
		name          string
		input         *QueryInput
		expectedError bool
	}{
		{
			name: "Valid Input",
			input: &QueryInput{
				payload: AddQueryRequest{
					Name:              "Test",
					Description:       "Description",
					Query:             "SELECT *",
					DefaultParameters: json.RawMessage(`{}`),
					DataProductID:     uuid.New(),
				},
			},
			expectedError: false,
		},
		{
			name: "Invalid Input 1",
			input: &QueryInput{
				payload: AddQueryRequest{
					Name:              "Test",
					Description:       "Description",
					Query:             "SELECT *",
					DefaultParameters: json.RawMessage(`{}`),
				},
			},
			expectedError: true,
		},
		{
			name: "Invalid Input 2",
			input: &QueryInput{
				payload: AddQueryRequest{
					Name:          "Test",
					Description:   "Description",
					Query:         "SELECT *",
					DataProductID: uuid.New(),
				},
			},
			expectedError: true,
		},
		{
			name: "Invalid Input 3",
			input: &QueryInput{
				payload: AddQueryRequest{
					Name:              "Test",
					Description:       "Description",
					DefaultParameters: json.RawMessage(`{}`),
					DataProductID:     uuid.New(),
				},
			},
			expectedError: true,
		},
		{
			name: "Invalid Input 4",
			input: &QueryInput{
				payload: AddQueryRequest{
					Name:              "Test",
					Query:             "SELECT *",
					DefaultParameters: json.RawMessage(`{}`),
					DataProductID:     uuid.New(),
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &application{}
			result := app.validateAddQueryRequestParameters(tt.input)
			assert.Equal(t, tt.expectedError, result)
		})
	}
}

func TestRunQuery(t *testing.T) {
	tests := []struct {
		name           string
		input          RunQueryRequest
		setupMock      func(*MockDB)
		expectedStatus int
		serverResponse interface{}
		statusCode     int
	}{
		{
			name: "Successful Query Execution",
			input: RunQueryRequest{
				QueryID:    uuid.New(),
				Parameters: json.RawMessage(`{}`),
			},
			setupMock: func(m *MockDB) {
				m.On("FetchQUeryString", mock.AnythingOfType("uuid.UUID")).Return("SELECT * FROM test", nil)
			},
			expectedStatus: http.StatusCreated,
			serverResponse: map[string]interface{}{
				"results": []map[string]interface{}{
					{
						"rows": []map[string]interface{}{
							{"id": 1, "name": "test"},
							{"id": 2, "name": "test2"},
						},
					},
				},
			},
			statusCode: 200,
		},
		{
			name: "Query Not Found",
			input: RunQueryRequest{
				QueryID: uuid.New(),
			},
			setupMock: func(m *MockDB) {
				m.On("FetchQUeryString", mock.AnythingOfType("uuid.UUID")).Return("", assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Data gateway error",
			input: RunQueryRequest{
				QueryID: uuid.New(),
			},
			setupMock: func(m *MockDB) {
				m.On("FetchQUeryString", mock.AnythingOfType("uuid.UUID")).Return("SELECT * FROM test", nil)
			},
			expectedStatus: http.StatusBadRequest,
			serverResponse: map[string]string{"error": "internal server error"},
			statusCode:     500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.setupMock(mockDB)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if str, ok := tt.serverResponse.(string); ok {
					w.Write([]byte(str))
				} else {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))

			newLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
			app := &application{
				db: mockDB,
				config: config{
					dataGatewayURL: server.URL,
				},
				logger: newLogger,
			}

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			app.RunQuey(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestFetchAllQueries(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockDB)
		expectedStatus int
	}{
		{
			name: "Successful Fetch",
			setupMock: func(m *MockDB) {
				m.On("FetchAllQUeries").Return([]database.Query{}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Database Error",
			setupMock: func(m *MockDB) {
				m.On("FetchAllQUeries").Return([]database.Query{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.setupMock(mockDB)

			newLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
			app := &application{db: mockDB, logger: newLogger}

			req := httptest.NewRequest(http.MethodGet, "/query", nil)
			rec := httptest.NewRecorder()

			app.FetchAllQueries(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockDB.AssertExpectations(t)
		})
	}
}
