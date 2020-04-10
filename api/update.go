package api

import (
	"database/sql"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type MetricsResponse struct {
	Metrics []MetricType `json:"metrics"`
}

func handleMetrics(db *sql.DB) (*MetricsResponse, *HttpError) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()

	res := MetricsResponse{Metrics: []MetricType{}}
	for _, m := range METRICS {
		res.Metrics = append(res.Metrics, *m)
	}
	return &res, nil
}

func AddMetrics(data Data) {
	prefix := []string(nil)
	name, err := GetString(data, "name")
	if err != nil {
		log.Warnf("Failed getting report type: %s", err)
	} else {
		prefix = append(prefix, name)
	}
	addMetricsJson(prefix, data)
}

func addMetric(prefix string, t string) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()

	prefix = I.I(prefix)
	t = I.I(t)

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
	} else if _, ok := v.(string); ok {
		addMetric(prefix, "string")
	} else if _, ok := v.(bool); ok {
		addMetric(prefix, "bool")
	} else if _, ok := v.(float64); ok {
		addMetric(prefix, "number")
	} else if v == nil {
		addMetric(prefix, "null")
	} else {
		log.Warnf("Not expected json type for value: %+v, prefix: %s", v, prefix)
	}
}

func AddData(user User, userId string, timestamp int64, data Data) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()
	USERS[userId] = user
	if _, ok := DATA[userId]; !ok {
		DATA[userId] = make(map[int64][]Data)
	}
	if _, ok := DATA[userId][timestamp]; !ok {
		DATA_SERIES[userId] = append(DATA_SERIES[userId], timestamp)
		sort.Slice(DATA_SERIES[userId], func(i, j int) bool { return DATA_SERIES[userId][i] < DATA_SERIES[userId][j] })
	}
	DATA[userId][timestamp] = append(DATA[userId][timestamp], data)
}

func ClearOld(userId string, timestamp int64) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()
	log.Infof("Before clean there are %d time series for %s", len(DATA_SERIES[userId]), userId)
	lastTimestamp := DATA_SERIES[userId][len(DATA_SERIES[userId])-1]
	log.Infof("firsTimestamp: %+v lastTimestamp: %+v", DATA_SERIES[userId][0], lastTimestamp)
	index := 0
	deleteTimestamp := DATA_SERIES[userId][index]
	for deleteTimestamp < lastTimestamp-STORE_TTL {
		delete(DATA[userId], timestamp)
		index++
		deleteTimestamp = DATA_SERIES[userId][index]
	}
	if index > 0 {
		DATA_SERIES[userId] = DATA_SERIES[userId][index:]
	}
	log.Infof("After clean there are %d time series for %s", len(DATA_SERIES[userId]), userId)
}

func handleUpdate(db *sql.DB, r map[string]interface{}) (*EmptyResponse, *HttpError) {
	user, err := GetUser(r)
	if err != nil {
		return &EmptyResponse{}, NewBadRequestError(err)
	}
	userId, err := UserId(user)
	if err != nil {
		return &EmptyResponse{}, NewBadRequestError(err)
	}
	datas, err := GetDatas(r)
	if err != nil {
		return &EmptyResponse{}, NewBadRequestError(err)
	}
	maxTimestamp := int64(0)
	for _, data := range datas {
		floatTimestamp, err := GetFloat64(data, "timestamp")
		if err != nil {
			return &EmptyResponse{}, NewBadRequestError(err)
		}
		timestamp := int64(floatTimestamp)
		AddData(user, userId, timestamp, data)
		if timestamp > maxTimestamp {
			maxTimestamp = timestamp
		}
		AddMetrics(data)
	}
	ClearOld(userId, maxTimestamp)

	return &EmptyResponse{}, nil
}
