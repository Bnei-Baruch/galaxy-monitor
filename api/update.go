package api

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//func AddData(user User, userId string, data Data) {
//	DATA_MUX.Lock()
//	defer DATA_MUX.Unlock()
//	if _, ok := DATA[userId][timestamp]; !ok {
//		DATA_SERIES[userId] = append(DATA_SERIES[userId], timestamp)
//		sort.Slice(DATA_SERIES[userId], func(i, j int) bool { return DATA_SERIES[userId][i] < DATA_SERIES[userId][j] })
//	}
//	DATA[userId][timestamp] = append(DATA[userId][timestamp], data)
//}

func AddUser(user User, userId string) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()
	USERS[userId] = user
	if _, ok := DATA[userId]; !ok {
		DATA[userId] = make(map[int64][]Data)
	}
}

func InternJson(v interface{}) interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return InternJsonMap(m)
	} else if a, ok := v.([]interface{}); ok {
		ia := make([]interface{}, 0)
		for i := range a {
			ia = append(ia, InternJson(a[i]))
		}
		return ia
	} else if s, ok := v.(string); ok {
		return I.I(s)
	} else {
		return v
	}
}

func InternJsonMap(v map[string]interface{}) map[string]interface{} {
	im := make(map[string]interface{})
	for k, v := range v {
		im[I.I(k)] = InternJson(v)
	}
	return im
}

func DInternJson(v interface{}) {
	if m, ok := v.(map[string]interface{}); ok {
		for k, v := range m {
			I.DI(k)
			DInternJson(v)
		}
	} else if a, ok := v.([]interface{}); ok {
		for i := range a {
			DInternJson(a[i])
		}
	} else if s, ok := v.(string); ok {
		I.DI(s)
	}
}

func DatasOnNameTimestamp(datasOnName []Data) (int64, error) {
	timestamp := int64(0)
	for i := range datasOnName {
		t, err := GetTimestamp(datasOnName[i])
		if err != nil {
			return 0, err
		}
		if timestamp == 0 {
			timestamp = t
		}
		// If time diff > 100 millis.
		if t-timestamp > 100 || timestamp-t > 100 {
			return 0, errors.New(fmt.Sprintf("Expected datas on name to have the same timestamp %d, got %d", timestamp, t))
		}
	}
	return timestamp, nil
}

func AddMetricData(userId string, datasOnName []Data) error {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()

	if _, ok := METRICS_DATA[userId]; !ok {
		METRICS_DATA[userId] = &MetricsData{
			Index:      make(map[string]int),
			Timestamps: []int64{},
			Data:       [][]interface{}{},
			Stats:      [][]*Stats{},
		}
	}

	timestamp, err := DatasOnNameTimestamp(datasOnName)
	if err != nil {
		return err
	}
	if timestamp == 0 {
		return errors.New(fmt.Sprintf("Timestamp is 0, probably no datas: %+v", datasOnName))
	}

	md := METRICS_DATA[userId]

	i := sort.Search(len(md.Timestamps),
		func(i int) bool { return md.Timestamps[i] >= timestamp })
	if i < len(md.Timestamps) && md.Timestamps[i] == timestamp {
		// timestamp is present at md.Timestamps[i]
		return nil
	} else {
		// timestamp is not present in data,
		// i is the index where it would be inserted.
		if i < len(md.Timestamps) {
			log.Warnf("Expected timestamp to be larger then %d, got %d. Ignoring.",
				md.Timestamps[len(md.Timestamps)-1], timestamp)
			return nil
		}
	}
	md.Timestamps = append(md.Timestamps, timestamp)

	for i, metric := range SPEC.MetricsWhitelist {
		if _, ok := md.Index[metric]; !ok {
			md.Index[metric] = len(md.Index)
		}
		parts := strings.Split(metric, ".")
		if len(md.Data) <= i {
			md.Data = append(md.Data, []interface{}{})
			md.Stats = append(md.Stats, []*Stats{NewStats(), NewStats(), NewStats()})
		}
		value := JsonMetric(datasOnName, parts)
		md.Data[i] = append(md.Data[i], value)
		if len(md.Data[i]) != len(md.Timestamps) {
			log.Errorf("Data length %d expected to be same as Timestamps length %d", len(md.Data[i]), len(md.Timestamps))
		}
		if valueFloat, ok := value.(float64); ok {
			for j := range md.Stats[i] {
				md.Stats[i][j].Add(valueFloat, timestamp)
			}
		}
	}

	return nil
}

func PrintJson(v interface{}) {
	str, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	log.Infof("Json: %s", str)
}

func JsonMetric(v interface{}, parts []string) interface{} {
	if len(parts) == 0 {
		return v
	}
	if strings.HasPrefix(parts[0], "[") && strings.HasSuffix(parts[0], "]") {
		fieldValue := strings.Split(parts[0][1:len(parts[0])-1], ":")
		if len(fieldValue) != 2 {
			log.Warnf("Failed extracting field and value from %s.", parts[0])
			return nil
		}
		field := fieldValue[0]
		value := fieldValue[1]
		a, ok := v.([]interface{})
		if !ok {
			am, ok := v.([]map[string]interface{})
			if !ok {
				log.Warnf("Did not find metric %s in value is not array it is %+v", parts[0], reflect.TypeOf(v))
				PrintJson(v)
				log.Infof("Parts: %+v", parts)
				return nil
			}
			a = []interface{}(nil)
			for i := range am {
				a = append(a, am[i])
			}
		}
		for i := range a {
			m, ok := a[i].(map[string]interface{})
			if !ok {
				log.Warnf("Failed exracting field from non map: %+v", a[i])
				PrintJson(a[i])
				log.Infof("Parts: %+v", parts)
				return nil
			}
			stringValue, err := GetString(m, field)
			if err != nil {
				log.Warnf("Field array filter %s not string or does not exist in %+v", parts[0], m)
				PrintJson(m)
				log.Infof("Parts: %+v", parts)
				return nil
			}
			if stringValue == value {
				return JsonMetric(a[i], parts[1:])
			}
		}
		// log.Warnf("Did not find %+v.", parts)
		// PrintJson(v)
		return nil
	}
	if m, ok := v.(map[string]interface{}); ok {
		value, found := m[parts[0]]
		if !found {
			//for key, value := range m {
			//	log.Infof("[%s] [%s]: %s key type %+v", key, parts[0], value, reflect.TypeOf(key))
			//}
			//log.Warnf("Failed finding field %s in map %+v.", parts[0], m)
			return nil
		}
		return JsonMetric(value, parts[1:])
	} else {
		log.Warnf("Expected map, got %+v for %s", v, parts[0])
		PrintJson(v)
		log.Infof("Parts: %+v", parts)
		return nil
	}
}

func ClearOldMetricsData(userId string) error {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()

	if _, ok := METRICS_DATA[userId]; !ok {
		return errors.New(fmt.Sprintf("No metrics data for %s", userId))
	}

	md := METRICS_DATA[userId]
	if len(md.Timestamps) == 0 {
		return nil
	}

	lastTimestamp := md.Timestamps[len(md.Timestamps)-1]

	timestampMinuteIndex := sort.Search(len(md.Timestamps),
		func(i int) bool { return md.Timestamps[i] >= lastTimestamp-MINUTE_MS })
	timestampThreeMinutesIndex := sort.Search(len(md.Timestamps),
		func(i int) bool { return md.Timestamps[i] >= lastTimestamp-THREE_MINUTES_MS })
	timestampAllIndex := sort.Search(len(md.Timestamps),
		func(i int) bool { return md.Timestamps[i] >= lastTimestamp-STORE_TTL })

	// Update stats.
	for _, metricIndex := range md.Index {
		statsIndices := []int{timestampMinuteIndex, timestampThreeMinutesIndex, timestampAllIndex}
		for statIndex, timestampIndex := range statsIndices {
			startStatIndex := sort.Search(len(md.Timestamps), func(i int) bool { return md.Timestamps[i] > md.Stats[metricIndex][statIndex].MaxRemovedTimestamp })
			for dataIndex := startStatIndex; dataIndex < timestampIndex; dataIndex++ {
				value := md.Data[metricIndex][dataIndex]
				if valueFloat, ok := value.(float64); ok {
					md.Stats[metricIndex][statIndex].Remove(valueFloat, md.Timestamps[dataIndex])
				}
			}
		}
	}

	md.Timestamps = md.Timestamps[timestampAllIndex:]
	for _, metricIndex := range md.Index {
		md.Data[metricIndex] = md.Data[metricIndex][timestampAllIndex:]

		if len(md.Timestamps) != len(md.Data[metricIndex]) {
			return errors.New(fmt.Sprintf("Expected timestamps length %d to be the same as data length %d", len(md.Timestamps), len(md.Data[metricIndex])))
		}
	}
	return nil
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
		// Delete interned strings
		for i := range DATA[userId][timestamp] {
			DInternJson(DATA[userId][timestamp][i])
		}
		delete(DATA[userId], timestamp)
		index++
		deleteTimestamp = DATA_SERIES[userId][index]
	}
	if index > 0 {
		DATA_SERIES[userId] = DATA_SERIES[userId][index:]
	}
	log.Infof("After clean there are %d time series for %s", len(DATA_SERIES[userId]), userId)
}

type UpdateResponse struct {
	Spec Spec `json:"spec"`
}

func handleUpdate(r map[string]interface{}) (*UpdateResponse, *HttpError) {
	user, err := GetUser(r)
	if err != nil {
		return &UpdateResponse{Spec: SPEC}, NewBadRequestError(err)
	}
	userId, err := UserId(user)
	if err != nil {
		return &UpdateResponse{Spec: SPEC}, NewBadRequestError(err)
	}

	datasOnTimestamp, err := GetDatas(r)
	if err != nil {
		return &UpdateResponse{Spec: SPEC}, NewBadRequestError(err)
	}
	for _, datasOnNames := range datasOnTimestamp {
		AddUser(user, userId)
		// internData := InternJsonMap(data)
		//AddData(userId, internData)
		AddMetrics(datasOnNames)
		err := AddMetricData(userId, datasOnNames)
		if err != nil {
			return nil, NewInternalError(err)
		}
	}
	// ClearOld(userId, maxTimestamp)
	ClearOldMetricsData(userId)

	return &UpdateResponse{Spec: SPEC}, nil
}
