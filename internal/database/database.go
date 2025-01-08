package database

import "github.com/google/uuid"

type Database interface {
	AddNewQueries(query Query) error
	FetchQUeryString(queryID uuid.UUID) (string, error)
	FetchAllQUeries() ([]Query, error)
}
