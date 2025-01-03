-- +goose Up
CREATE TABLE queries(
    query_id UUID  PRIMARY KEY,
    name VARCHAR(40),
    data_product_id UUID,
    description TEXT,
    query TEXT,
    default_parameters jsonb
);

-- +goose Down
DROP TABLE queries;

-- goose postgres postgres://postgres:pass@localhost:5432/postgres up