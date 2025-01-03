package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"xcaliber/data-quality-metrics-framework/internal/database"
	"xcaliber/data-quality-metrics-framework/internal/env"
	"xcaliber/data-quality-metrics-framework/internal/metrics"
	"xcaliber/data-quality-metrics-framework/internal/version"
	"xcaliber/data-quality-metrics-framework/internal/workflow"

	"github.com/prometheus/client_golang/prometheus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
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
	temporalHost         string
	temporalPort         int
	temporalTaskQueue    string
	temporalCronSchedule string
	dataGatewayURL       string
}

type application struct {
	config config
	db     database.Database
	logger *slog.Logger
	wg     sync.WaitGroup
}

func init() {
	prometheus.MustRegister(metrics.QueryOutput)
}

func startWorkflowScheduler(cfg config, c client.Client, twf *workflow.TemporalWorkflow, cronSchedule string) {

	options := client.StartWorkflowOptions{
		ID:           "data-quality-metric-framework-runQuery-CronSchedule-workflow",
		TaskQueue:    cfg.temporalTaskQueue,
		CronSchedule: cronSchedule,
	}

	_, err := c.ExecuteWorkflow(context.Background(), options, twf.MyWorkflow)
	if err != nil {
		fmt.Println("Failed to start workflow", err)
	}
}

func startWorker(cfg config, c client.Client, act *workflow.TemporalWorkflow) {
	// create worker for task queue
	w := worker.New(c, cfg.temporalTaskQueue, worker.Options{})

	// register workflow and activity
	w.RegisterWorkflow(act.MyWorkflow)
	w.RegisterActivity(act.MyActivity)
	err := w.Run(worker.InterruptCh())
	if err != nil {
		fmt.Println("Unable to start worker", err)
	}
}

func run(logger *slog.Logger) error {
	var cfg config

	cfg.baseURL = env.GetString("BASE_URL", "http://0.0.0.0:4444")
	cfg.host = env.GetString("HOST", "0.0.0.0")
	cfg.httpPort = env.GetInt("HTTP_PORT", 4444)
	cfg.db.dsn = env.GetString("DB_CONNECTION_STRING", "postgres:pass@localhost:5432/postgres?sslmode=disable")
	cfg.temporalHost = env.GetString("TEMPORAL_HOST", "localhost")
	cfg.temporalPort = env.GetInt("TEMPORAL_PORT", 7233)
	cfg.temporalTaskQueue = env.GetString("TEMPORAL_TASK_QUEUE", "data_quality_metrics")
	cfg.temporalCronSchedule = env.GetString("TEMPORAL_CRON_SCHEDULE", "0 0 * * *") // Once every minute
	cfg.dataGatewayURL = env.GetString("DATA_GATEWAY_URL", "https://blitz.xcaliberapis.com/xcaliber-dev/gateway/api/v2/query/rows")

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

	// temporal client
	c, err := client.Dial(client.Options{
		HostPort: fmt.Sprintf("%s:%d", cfg.temporalHost, cfg.temporalPort),
	})
	if err != nil {
		fmt.Println("Unable to create Temporal client", err)
	}
	defer c.Close()

	logger.Info("Temporal client started successfully")

	twf := &workflow.TemporalWorkflow{
		DB:             db,
		DataGatewayURL: cfg.dataGatewayURL,
		Logger:         logger,
	}
	go startWorkflowScheduler(cfg, c, twf, cfg.temporalCronSchedule)
	go startWorker(cfg, c, twf)

	return app.serveHTTP()
}
