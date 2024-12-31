-- +goose Up
CREATE TABLE queries(
    query_id UUID  PRIMARY KEY,
    name VARCHAR(40),
    description TEXT,
    query TEXT 
);

-- +goose Down
DROP TABLE queries;

-- goose postgres postgres://postgres:pass@localhost:5432/postgres up