package api

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type MetricsResponse struct {
	Metrics []MetricType `json:"metrics"`
}

func handleMetrics() (*MetricsResponse, *HttpError) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()

	res := MetricsResponse{Metrics: []MetricType{}}
	for _, m := range METRICS {
		res.Metrics = append(res.Metrics, *m)
	}
	return &res, nil
}

func AddMetrics(data []Data) {
	addMetricsJson([]string{}, data)
}

func addMetric(prefix string, t string) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()

	if _, ok := METRICS[prefix]; !ok {
		METRICS[prefix] = &MetricType{Metric: prefix, Type: t, Updates: 0}
	}
	if METRICS[prefix].Type == "null" {
		METRICS[prefix].Type = t
	}
	if t != "null" && METRICS[prefix].Type != t {
		log.Warnf("Metric %s expected to be %s, but is %s.", prefix, t, METRICS[prefix].Type)
	}
	METRICS[prefix].Updates++
}

func addMetricsJson(path []string, v interface{}) {
	prefix := strings.Join(path, ".")
	if m, ok := v.(map[string]interface{}); ok {
		if len(path) != 0 {
			addMetric(prefix, "map")
		}
		for k, v := range m {
			newPath := append([]string(nil), path...)
			if len(path) >= 2 && path[len(path)-1] == "[]" && path[len(path)-2] == "reports" {
				reportType, err := GetString(m, "type")
				if err == nil {
					newPath = append(newPath, fmt.Sprintf("[type:%s]", reportType))
				}
			}
			if len(path) == 0 || (len(path) == 1 && path[0] == "[]") {
				dataName, err := GetString(m, "name")
				if err == nil {
					newPath = append(newPath, fmt.Sprintf("[name:%s]", dataName))
				}
			}
			newPath = append(newPath, k)
			addMetricsJson(newPath, v)
		}
	} else if a, ok := v.([]interface{}); ok {
		if len(path) != 0 {
			addMetric(prefix, "array")
		}
		for i := range a {
			newPath := append([]string(nil), path...)
			newPath = append(newPath, "[]")
			addMetricsJson(newPath, a[i])
		}
	} else if a, ok := v.([]map[string]interface{}); ok {
		if len(path) != 0 {
			addMetric(prefix, "array")
		}
		for i := range a {
			newPath := append([]string(nil), path...)
			newPath = append(newPath, "[]")
			addMetricsJson(newPath, a[i])
		}
	} else if _, ok := v.(string); ok {
		addMetric(prefix, "string")
	} else if _, ok := v.(bool); ok {
		addMetric(prefix, "bool")
	} else if _, ok := v.(float64); ok {
		addMetric(prefix, "number")
	} else if v == nil {
		addMetric(prefix, "null")
	} else {
		log.Warnf("Not expected json type for prefix: %s type: %+v", prefix, reflect.TypeOf(v))
		PrintJson(v)
	}
}
