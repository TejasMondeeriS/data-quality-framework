package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var QueryOutput = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "query_output",
		Help: "Sets the result for every query.",
	},
	[]string{"name", "data_product_id"},
)

func SetMetricValue(name string, value float64, data_product_id string) {
	QueryOutput.With(prometheus.Labels{"name": name, "data_product_id": data_product_id}).Set(value)
}
