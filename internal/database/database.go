package database

import "github.com/google/uuid"

type Database interface {
	AddNewQueries(query Query) error
	FetchQUeryString(queryID uuid.UUID) (string, error)
	// RunQuery(query string) ([]map[string]interface{}, error)
	FetchAllQUeries() ([]Query, error)
}
