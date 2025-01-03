package workflow

import (
	"context"
	"log/slog"
	"time"
	datagateway "xcaliber/data-quality-metrics-framework/internal/data_gateway"
	"xcaliber/data-quality-metrics-framework/internal/database"
	"xcaliber/data-quality-metrics-framework/internal/metrics"
	"xcaliber/data-quality-metrics-framework/internal/utility"

	"go.temporal.io/sdk/workflow"
)

type TemporalWorkflow struct {
	DB             database.Database
	DataGatewayURL string
	Logger         *slog.Logger
}

func (twf *TemporalWorkflow) MyWorkflow(ctx workflow.Context) error {
	// activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 1,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// execute activity
	err := workflow.ExecuteActivity(ctx, twf.MyActivity).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (twf *TemporalWorkflow) MyActivity(ctx context.Context) error {
	queries, err := twf.DB.FetchAllQUeries()
	if err != nil {
		return err
	}

	for _, query := range queries {
		formattedQuery, err := utility.FormatQuery(query.Query, query.DefaultParameters)
		if err != nil {
			return err
		}
		results, err := datagateway.RunQuery(twf.DataGatewayURL, formattedQuery)
		if err != nil {
			twf.Logger.Error("Error while running query: %v: %v", slog.Any("name", query.Name), slog.Any("err", err))
			continue
		}
		if len(results) > 1 || len(results[0]) > 1 {
			twf.Logger.Error("query does not return a single value, returns multiple rows or columns: %v", slog.Any("name", query.Name))
			continue
		}
		res := 0.0
		for _, v := range results[0] {
			switch y := v.(type) {
			case int, int8, int16, int32, int64:
				res = float64(y.(int64))
			case uint, uint8, uint16, uint32, uint64:
				res = float64(y.(uint64))
			case float32:
				res = float64(y)
			case float64:
				res = y
			default:
				twf.Logger.Error("query returns non-numeric type", slog.Any("name", query.Name), slog.Any("value", v))
			}

		}
		metrics.SetMetricValue(query.Name, res, query.DataProductID.String())
		twf.Logger.Info("query ran successfully: %v, %v", slog.Any("name", query.Name), slog.Any("value", res))
	}

	return nil

}
