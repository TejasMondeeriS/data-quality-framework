package database

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
)

type Query struct {
	QueryID     uuid.UUID `json:"query_id" db:"query_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Query       string    `json:"query" db:"query"`
}

func (db *DB) AddNewQueries(query Query) error {
	sb := db.builder.Flavor.NewInsertBuilder()
	sb.InsertInto("Queries")
	sb.Cols(
		"query_id", "name", "description", "query",
	)

	sb.Values(
		query.QueryID, query.Name, query.Description, query.Query,
	)

	sql, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	_, err := db.Exec(sql, args...)
	return err
}

func (db *DB) FetchQUery(queryID uuid.UUID) (string, error) {
	sb := db.builder.Flavor.NewSelectBuilder()
	sb.Select("query")
	sb.From("Queries")
	sb.Where(sb.Equal("query_id", queryID))
	sql, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	query := Query{}
	err := db.Get(&query, sql, args...)
	return query.Query, err
}

func (db *DB) RunQuery(query string) ([]map[string]interface{}, error) {
	rows, err := db.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		err := rows.MapScan(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, row)
	}

	return results, nil

}
