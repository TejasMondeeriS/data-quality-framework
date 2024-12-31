package database

import (
	"context"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

const defaultTimeout = 3 * time.Second

type DB struct {
	*sqlx.DB
	builder *sqlbuilder.Struct
}

func New(dsn string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "postgres", "postgres://"+dsn)

	sslStatus := ""
	db.QueryRow("SHOW ssl").Scan(&sslStatus)
	fmt.Println("SSL Mode:", sslStatus)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(2 * time.Hour)

	builder := sqlbuilder.NewStruct(new(struct{}))

	return &DB{DB: db, builder: builder}, nil
}
