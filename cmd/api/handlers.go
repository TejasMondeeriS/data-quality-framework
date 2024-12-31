package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"xcaliber/data-quality-metrics-framework/internal/database"
	"xcaliber/data-quality-metrics-framework/internal/request"
	"xcaliber/data-quality-metrics-framework/internal/response"
	"xcaliber/data-quality-metrics-framework/internal/validator"

	"github.com/google/uuid"
)

// Health godoc
// @Summary Service Health endpoint
// @Description Endpoint to determine the health of the service
// @Tags health
// @Produce  json
// @Success 201 {object} map[string]string "{"Status": "OK"}"
// @Failure 500 {object} map[string]string "{"error": "Internal server error"}"
// @Router /health [get]
func (app *application) HealthHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Status": "OK",
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

type QueryInput struct {
	payload   AddQueryRequest
	Validator validator.Validator `json:"-"`
}

// Add a new Query
// @Summary Add a new Query
// @Description Endpoint to Add/Register a new Query
// @Tags query
// @Produce  json
// @Success 201 {object} map[string]string "{"Data":map[string]interface{},"Status": "OK", "Message":"Query added successfully"}"
// @Failure 400 {object} map[string]string "{"error": "invalid request"}"
// @Failure 500 {object} map[string]string "{"error": "Internal server error"}"
// @Router /query [post]
func (app *application) AddQuery(w http.ResponseWriter, r *http.Request) {
	var input QueryInput
	err := request.DecodeJSON(w, r, &input.payload)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// Validate input
	hasErrors := app.validateAddQueryRequestParameters(&input)
	if hasErrors {
		app.failedValidation(w, r, input.Validator)
		return
	}

	query := database.Query{
		QueryID:     uuid.New(),
		Name:        input.payload.Name,
		Description: input.payload.Description,
		Query:       input.payload.Query,
	}
	err = app.db.AddNewQueries(query)
	if err != nil {
		app.logger.Error("Error adding new query: %s", slog.Any("err", err))
		app.serverError(w, r, errors.New("error inseting query"))
		return
	}

	// Response
	res := StandardResponse{
		Status:  http.StatusText(http.StatusCreated),
		Message: "Query added successfully",
		Data: struct {
			QueryID uuid.UUID `json:"query_id"`
		}{query.QueryID},
	}
	err = response.JSON(w, http.StatusCreated, res)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) validateAddQueryRequestParameters(
	input *QueryInput,
) bool {

	input.Validator.CheckField(
		input.payload.Name != "",
		"Name",
		"Name is required",
	)
	input.Validator.CheckField(
		input.payload.Description != "",
		"Description",
		"Description is required",
	)
	input.Validator.CheckField(
		input.payload.Query != "",
		"Query",
		"Query is required",
	)

	return input.Validator.HasErrors()

}

type RunQueryInput struct {
	payload   RunQueryRequest
	Validator validator.Validator `json:"-"`
}

// Run stored query
// @Summary Run stored query
// @Description Endpoint to Run a stored query
// @Tags run
// @Produce  json
// @Success 201 {object} map[string]string "{"Data":map[string]interface{},"Status": "OK", "Message":"Query added successfully"}"
// @Failure 400 {object} map[string]string "{"error": "invalid request"}"
// @Failure 500 {object} map[string]string "{"error": "Internal server error"}"
// @Router /run [post]
func (app *application) RunQuey(w http.ResponseWriter, r *http.Request) {
	var input RunQueryInput
	err := request.DecodeJSON(w, r, &input.payload)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	query, err := app.db.FetchQUery(input.payload.QueryID)
	if err != nil {
		app.logger.Error("Error fetching query: %s", slog.Any("err", err))
		app.serverError(w, r, errors.New("error fetching query"))
		return
	}

	for key, val := range input.payload.Parameters {
		placeholder := "$" + key
		var replacement string
		switch v := val.(type) {
		case string:
			if strings.HasPrefix(v, "now()") { //  timestamp-like string
				replacement, err = app.parseTimestamp(v)
				if err != nil {
					app.badRequest(w, r, fmt.Errorf("invalid timestamp format for %v: %v", key, val))
					return
				}
			} else { // regular string
				replacement = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
			}
		case int, int8, int16, int32, int64:
			replacement = fmt.Sprintf("%d", v)
		case float32, float64:
			replacement = fmt.Sprintf("%f", v)
		case bool:
			replacement = strconv.FormatBool(v)
		default:
			app.badRequest(w, r, fmt.Errorf("unsupported parameter type for %v: %v", key, val))
			return
		}
		query = strings.ReplaceAll(query, placeholder, replacement)
	}

	results, err := app.db.RunQuery(query)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	res := StandardResponse{
		Status:  http.StatusText(http.StatusOK),
		Message: "Query executed successfully",
		Data:    results,
	}
	err = response.JSON(w, http.StatusCreated, res)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) parseTimestamp(input string) (string, error) {
	now := time.Now()

	if input == "now()" {
		return fmt.Sprintf("'%s'", now.Format("2006-01-02 15:04:05")), nil
	}

	if strings.HasPrefix(input, "now()-") {
		durationStr := strings.TrimPrefix(input, "now()-")
		timestamp := time.Now()

		if strings.HasSuffix(durationStr, "d") {
			days, err := strconv.Atoi(strings.TrimSuffix(durationStr, "d"))
			if err != nil {
				return "", fmt.Errorf("invalid day duration: %v", err)
			}
			timestamp = timestamp.AddDate(0, 0, -days)
		} else if strings.HasSuffix(durationStr, "w") {
			weeks, err := strconv.Atoi(strings.TrimSuffix(durationStr, "w"))
			if err != nil {
				return "", fmt.Errorf("invalid week duration: %v", err)
			}
			timestamp = timestamp.AddDate(0, 0, -7*weeks)
		} else if strings.HasSuffix(durationStr, "y") {
			years, err := strconv.Atoi(strings.TrimSuffix(durationStr, "y"))
			if err != nil {
				return "", fmt.Errorf("invalid year duration: %v", err)
			}
			timestamp = timestamp.AddDate(0, 0, -365*years)
		} else {
			duration, err := time.ParseDuration(strings.ReplaceAll(durationStr, " ", ""))
			if err != nil {
				return "", fmt.Errorf("invalid duration format: %v", err)
			}
			timestamp = timestamp.Add(-duration)
		}

		return fmt.Sprintf("'%s'", timestamp.Format("2006-01-02 15:04:05")), nil
	}
	return "", fmt.Errorf("unsupported timestamp format")
}
