package main

import (
	"github.com/google/uuid"
)

type AddQueryRequest struct {
	Name        string `json:"name"        binding:"required"`
	Query       string `json:"query"       binding:"required"`
	Description string `json:"description" binding:"required"`
}

type RunQueryRequest struct {
	QueryID    uuid.UUID              `json:"query_id"   binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
}
