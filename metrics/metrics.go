package metrics

import (
	"fmt"
	"strings"

	"github.com/rcrowley/go-metrics"
)

func makeMetricName(measurement string, tags ...string) string {
	result := []string{measurement}
	for i := 0; i < len(tags); i += 2 {
		result = append(result, fmt.Sprintf("%s=%s", tags[i], tags[i+1]))
	}
	return strings.Join(result, ",")
}

func getMeasurementAndTags(name string) (string, map[string]string) {
	measurementAndTags := strings.Split(name, ",")
	tags := map[string]string{}
	for i := 1; i < len(measurementAndTags); i++ {
		kvs := strings.Split(measurementAndTags[i], "=")
		tags[kvs[0]] = kvs[1]
	}
	return measurementAndTags[0], tags
}

func NewCounter(measurement string, tags ...string) metrics.Counter {
	return metrics.GetOrRegisterCounter(makeMetricName(measurement, tags...), nil)
}

func NewGauge(measurement string, tags ...string) metrics.Gauge {
	return metrics.GetOrRegisterGauge(makeMetricName(measurement, tags...), nil)
}

func NewGaugeFloat64(measurement string, tags ...string) metrics.GaugeFloat64 {
	return metrics.GetOrRegisterGaugeFloat64(makeMetricName(measurement, tags...), nil)
}

func NewHistogram(measurement string, tags ...string) metrics.Histogram {
	return metrics.GetOrRegisterHistogram(makeMetricName(measurement, tags...), nil, metrics.NewExpDecaySample(1028, 0.015))
}

func NewMeter(measurement string, tags ...string) metrics.Meter {
	return metrics.GetOrRegisterMeter(makeMetricName(measurement, tags...), nil)
}

func NewTimer(measurement string, tags ...string) metrics.Timer {
	return metrics.GetOrRegisterTimer(makeMetricName(measurement, tags...), nil)
}
