# Data Quality Metrics Framework

A framework to allow users to run parameterized sql queries.

## ENV variables
### required:
HOST
HTTP_PORT
TEMPORAL_HOST
TEMPORAL_PORT
DATA_GATEWAY_URL
## optional:
TEMPORAL_TASK_QUEUE    | Default: data_quality_metrics

## Setup dev enviornment:

To setup dev enviornment, cd into ``` ./assets/dev_env ``` and run ``` docker-compose up -d ```

## Run the application:
``` go run ./cmd/api ```

Note: Currently prometheus scrapes every 15s, can be changed in ``` assets/dev_env/prometheus.yml ```
