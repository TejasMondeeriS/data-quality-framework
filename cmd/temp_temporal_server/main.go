package main

import (
	"context"
	"encoding/json"
	"fmt"
	"xcaliber/data-quality-metrics-framework/internal/database"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort: fmt.Sprintf("%s:%d", "localhost", 7233),
	})
	if err != nil {
		fmt.Println("Unable to create Temporal client", err)
		return
	}
	defer c.Close()

	fmt.Println("Temporal server started successfully")

	startWorkflowScheduler(c)

}

func startWorkflowScheduler(c client.Client) {

	options := client.StartWorkflowOptions{
		ID:           "data-quality-metric-framework-runQuery-CronSchedule-workflow",
		TaskQueue:    "data_quality_metrics",
		CronSchedule: "* * * * *",
	}

	query := database.Query{
		DataProductID: uuid.New(),
		Name:          "Test1",
		Query:         "select count(*) from xho_data_patient where birth_datetime > $timestamp;",
		Parameters:    json.RawMessage(`{"timestamp":"now()-30y"}`),
	}

	queryJson, err := json.Marshal(&query)
	if err != nil {
		fmt.Printf("Error marshalling query: %v\n", err)
		return
	}

	_, err = c.ExecuteWorkflow(context.Background(), options, "MyWorkflow", queryJson)
	if err != nil {
		fmt.Println("Failed to start workflow", err)
	}
}
