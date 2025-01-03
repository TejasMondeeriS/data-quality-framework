package database

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Query struct {
	QueryID           uuid.UUID       `json:"query_id"           db:"query_id"`
	DataProductID     uuid.UUID       `json:"data_product_id"    db:"data_product_id"`
	Name              string          `json:"name"               db:"name"`
	Description       string          `json:"description"        db:"description"`
	Query             string          `json:"query"              db:"query"`
	DefaultParameters json.RawMessage `json:"default_parameters" db:"default_parameters"`
}
