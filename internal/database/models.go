package database

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Query struct {
	DataProductID uuid.UUID       `json:"data_product_id"    db:"data_product_id"`
	Name          string          `json:"name"               db:"name"`
	Query         string          `json:"query"              db:"query"`
	Parameters    json.RawMessage `json:"parameters" db:"parameters"`
}
