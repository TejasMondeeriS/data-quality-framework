package main

import (
	"net/http"
	datagateway "xcaliber/data-quality-metrics-framework/internal/data_gateway"
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

func (app *application) validateRunQueryRequestParameters(
	input *RunQueryInput,
) bool {

	input.Validator.CheckField(
		input.payload.Name != "",
		"Name",
		"Name is required",
	)
	input.Validator.CheckField(
		input.payload.Query != "",
		"Query",
		"Query is required",
	)
	input.Validator.CheckField(
		input.payload.Parameters != nil,
		"Parameters",
		"Parameters is required",
	)
	input.Validator.CheckField(
		input.payload.DataProductID != uuid.Nil,
		"DataProductID",
		"DataProductID is required",
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

	ok := app.validateRunQueryRequestParameters(&input)
	if !ok {
		app.failedValidation(w, r, input.Validator)
		return
	}

	queryStr, err := utility.FormatQuery(input.payload.Query, input.payload.Parameters)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	results, err := datagateway.RunQuery(app.config.dataGatewayURL, queryStr)
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
