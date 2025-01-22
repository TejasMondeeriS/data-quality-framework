package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

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

// @host localhost:4446
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
	host              string
	httpPort          int
	temporalHost      string
	temporalPort      int
	temporalTaskQueue string
	dataGatewayURL    string
	temporalNamespace string
}

type application struct {
	config config
	logger *slog.Logger
	wg     sync.WaitGroup
}

func init() {
	prometheus.MustRegister(metrics.QueryOutput)
}

func startWorker(cfg config, c client.Client, act *workflow.TemporalWorkflow) {
	// create worker for task queue
	w := worker.New(c, cfg.temporalTaskQueue, worker.Options{})

	// register workflow and activity
	w.RegisterWorkflow(act.RunQueryWorkflow)
	w.RegisterActivity(act.RunQueryActivity)
	err := w.Run(worker.InterruptCh())
	if err != nil {
		fmt.Println("Unable to start worker", err)
	}
}

func run(logger *slog.Logger) error {
	var cfg config

	cfg.host = env.GetString("HOST", "0.0.0.0")
	cfg.httpPort = env.GetInt("HTTP_PORT", 4446)
	cfg.temporalHost = env.GetString("TEMPORAL_HOST", "localhost")
	cfg.temporalPort = env.GetInt("TEMPORAL_PORT", 7233)
	cfg.temporalTaskQueue = env.GetString("TEMPORAL_TASK_QUEUE", "data_quality_metrics")
	cfg.temporalNamespace = env.GetString("TEMPORAL_NAMESPACE", "testNamespace1")
	cfg.dataGatewayURL = env.GetString("DATA_GATEWAY_URL", "https://blitz.xcaliberapis.com/xcaliber-dev/gateway/api/v2/query/rows")

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	app := &application{
		config: cfg,
		logger: logger,
	}

	// temporal client
	c, err := client.Dial(client.Options{
		HostPort:  fmt.Sprintf("%s:%d", cfg.temporalHost, cfg.temporalPort),
		Namespace: cfg.temporalNamespace,
	})
	if err != nil {
		fmt.Println("Unable to create Temporal client", err)
	}
	defer c.Close()

	logger.Info("Temporal client started successfully")

	twf := &workflow.TemporalWorkflow{
		DataGatewayURL: cfg.dataGatewayURL,
		Logger:         logger,
	}
	// go startWorkflowScheduler(cfg, c, twf, cfg.temporalCronSchedule)
	go startWorker(cfg, c, twf)

	return app.serveHTTP()
}
