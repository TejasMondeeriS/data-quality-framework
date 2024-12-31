package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"xcaliber/data-quality-metrics-framework/internal/database"
	"xcaliber/data-quality-metrics-framework/internal/env"
	"xcaliber/data-quality-metrics-framework/internal/version"
)

// @title Data Quality Metrics Framework
// @version 1.0
// @description Used to monitor data quality

// @host localhost:4444
// @BasePath /
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type config struct {
	baseURL  string
	host     string
	httpPort int
	db       struct {
		dsn         string
		automigrate bool
	}
}

type application struct {
	config config
	db     database.Database
	logger *slog.Logger
	wg     sync.WaitGroup
}

func run(logger *slog.Logger) error {
	var cfg config

	cfg.baseURL = env.GetString("BASE_URL", "http://0.0.0.0:4444")
	cfg.host = env.GetString("HOST", "0.0.0.0")
	cfg.httpPort = env.GetInt("HTTP_PORT", 4444)
	cfg.db.dsn = env.GetString("DB_DSN", "postgres:pass@localhost:5432/postgres?sslmode=disable")

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	db, err := database.New(cfg.db.dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	app := &application{
		config: cfg,
		db:     db,
		logger: logger,
	}

	return app.serveHTTP()
}
