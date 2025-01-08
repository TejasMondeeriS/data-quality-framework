# Data Quality Metrics Framework

A framework to allow users to store parameterized sql queries and run them with arguments.

Note: Currently the framework is not integrated with control plane

## ENV variables
### required:
HOST
HTTP_PORT
DB_CONNECTION_STRING
TEMPORAL_HOST
TEMPORAL_PORT
DATA_GATEWAY_URL
## optional:
TEMPORAL_TASK_QUEUE    | Default: data_quality_metrics
TEMPORAL_CRON_SCHEDULE | Default : * * * * * -> runs every minute

## Setup dev enviornment:

To setup dev enviornment, cd into ``` ./assets/dev_env ``` and run ``` docker-compose up -d ```

To run migrations, install goose:
``` go install github.com/pressly/goose/v3/cmd/goose@latest  ```
cd into ./assets/schema and run goose on schema file:
``` goose postgres postgres://postgres:pass@localhost:5432/postgres up ```

## Run the application:
``` go run ./cmd/api ```

Note: Currently prometheus scrapes every 15s, can be changed in ``` assets/dev_env/prometheus.yml ```
