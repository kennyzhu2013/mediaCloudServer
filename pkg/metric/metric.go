package metric

import "common/metrics"

// TODO.
func InitMetrics(id, ver string) {
	// init metric here
	metrics.DefaultPrometheus = metrics.NewGinHandlerWrapper(metrics.Namespace("xmediaEmu"),
		metrics.Id(id), metrics.Version(ver),)
}
