package datagateway_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	datagateway "xcaliber/data-quality-metrics-framework/internal/data_gateway"
)

func TestRunQuery(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		serverResponse interface{}
		statusCode     int
		wantErr        bool
		expectedRows   int
	}{
		{
			name:  "successful query",
			query: "SELECT * FROM table",
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
			statusCode:   200,
			wantErr:      false,
			expectedRows: 2,
		},
		{
			name:           "server error",
			query:          "SELECT * FROM table",
			serverResponse: map[string]string{"error": "internal server error"},
			statusCode:     500,
			wantErr:        true,
			expectedRows:   0,
		},
		{
			name:           "invalid json response",
			query:          "SELECT * FROM table",
			serverResponse: "invalid json",
			statusCode:     200,
			wantErr:        true,
			expectedRows:   0,
		},
		{
			name:  "empty response",
			query: "SELECT * FROM empty_table",
			serverResponse: map[string]interface{}{
				"results": []map[string]interface{}{
					{
						"rows": []map[string]interface{}{},
					},
				},
			},
			statusCode:   200,
			wantErr:      false,
			expectedRows: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST request, got %s", r.Method)
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
				}

				var reqBody map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Errorf("failed to decode request body: %v", err)
					return
				}
				if sql, ok := reqBody["sql"].(string); ok {
					if sql != tt.query {
						t.Errorf("expected query %q, got %q", tt.query, sql)
						return
					}
				} else {
					t.Error("sql field not found in request body or not a string")
					return
				}

				w.WriteHeader(tt.statusCode)
				if str, ok := tt.serverResponse.(string); ok {
					w.Write([]byte(str))
				} else {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			rows, err := datagateway.RunQuery(server.URL, tt.query)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				expectedRows, ok := tt.serverResponse.(map[string]interface{})["results"].([]map[string]interface{})[0]["rows"].([]map[string]interface{})
				if !ok {
					t.Fatal("could not parse expected rows from server response")
				}

				if len(rows) != len(expectedRows) {
					t.Errorf("RunQuery() returned %d rows, expected %d", len(rows), len(expectedRows))
					return
				}

			}
		})
	}
}

func TestRunQueryInvalidURL(t *testing.T) {
	_, err := datagateway.RunQuery("invalid-url", "SELECT * FROM table")
	if err == nil {
		t.Error("RunQuery() expected error for invalid URL, got nil")
	}
	if x := err.Error(); !strings.Contains(x, "error making POST request") {
		t.Errorf("Received different error: %v", err)
	}
}
