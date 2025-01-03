package main

import (
	"errors"
	"log/slog"
	"net/http"
	datagateway "xcaliber/data-quality-metrics-framework/internal/data_gateway"
	"xcaliber/data-quality-metrics-framework/internal/database"
	"xcaliber/data-quality-metrics-framework/internal/request"
	"xcaliber/data-quality-metrics-framework/internal/response"
	"xcaliber/data-quality-metrics-framework/internal/utility"
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
		QueryID:           uuid.New(),
		Name:              input.payload.Name,
		Description:       input.payload.Description,
		Query:             input.payload.Query,
		DefaultParameters: input.payload.DefaultParameters,
		DataProductID:     input.payload.DataProductID,
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
	input.Validator.CheckField(
		input.payload.DefaultParameters != nil,
		"Default Parameters",
		"Default Parameters is required",
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
// @Success 201 {object} map[string]string "{"Data":map[string]interface{},"Status": "OK", "Message":"Query executed successfully"}"
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

	query, err := app.db.FetchQUeryString(input.payload.QueryID)
	if err != nil {
		app.logger.Error("Error fetching query: %s", slog.Any("err", err))
		app.serverError(w, r, errors.New("error fetching query"))
		return
	}

	query, err = utility.FormatQuery(query, input.payload.Parameters)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	results, err := datagateway.RunQuery(app.config.dataGatewayURL, query)
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

// Fetch All stored queries
// @Summary Fetch All stored queries
// @Description Endpoint to fetch All stored queries
// @Tags query
// @Produce  json
// @Success 201 {object} map[string]string "{"Data":map[string]interface{},"Status": "OK", "Message":"Queries fetched successfully"}"
// @Failure 500 {object} map[string]string "{"error": "Internal server error"}"
// @Router /query [get]
func (app *application) FetchAllQueries(w http.ResponseWriter, r *http.Request) {
	results, err := app.db.FetchAllQUeries()
	if err != nil {
		app.logger.Error("Error fetching queries: %v", slog.Any("err", err))
		app.serverError(w, r, errors.New("error fetching queries"))
		return
	}

	res := StandardResponse{
		Status:  http.StatusText(http.StatusOK),
		Message: "Queries fetched successfully",
		Data:    results,
	}
	err = response.JSON(w, http.StatusCreated, res)
	if err != nil {
		app.serverError(w, r, err)
	}
}
