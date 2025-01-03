package main

import (
	"encoding/json"

	"github.com/google/uuid"
)

type AddQueryRequest struct {
	Name              string          `json:"name"               binding:"required"`
	DataProductID     uuid.UUID       `json:"data_product_id"    binding:"required"`
	Query             string          `json:"query"              binding:"required"`
	Description       string          `json:"description"        binding:"required"`
	DefaultParameters json.RawMessage `json:"default_parameters" binding:"required"`
}

type RunQueryRequest struct {
	QueryID    uuid.UUID       `json:"query_id"   binding:"required"`
	Parameters json.RawMessage `json:"parameters"`
}
