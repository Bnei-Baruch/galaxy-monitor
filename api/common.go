package api

import (
	"fmt"
	"reflect"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

const (
	SECOND_MS        = 1000
	FIVE_SECONDS_MS  = 5 * SECOND_MS
	MINUTE_MS        = 60 * SECOND_MS
	TWO_MINUTES_MS   = 2 * MINUTE_MS
	THREE_MINUTES_MS = 3 * MINUTE_MS
	FIVE_MINUTES_MS  = 5 * MINUTE_MS
	TEN_MINUTES_MS   = 10 * MINUTE_MS
)

type Data = map[string]interface{}
type User = map[string]interface{}
type MetricType struct {
	Metric  string `json:"metric"`
	Type    string `json:"type"`
	Updates int64  `json:"updates"`
}
type Spec struct {
	SampleInterval   int64    `json:"sample_interval"`
	StoreInterval    int64    `json:"store_interval"`
	MetricsWhitelist []string `json:"metrics_whitelist"`
}
type MetricsData struct {
	Index      map[string]int  `json:"index"`
	Timestamps []int64         `json:"timestamps"`
	Data       [][]interface{} `json:"data"`
	Stats      [][]*Stats      `json:"stats"`
}

var (
	// Map from user to timestamp to data.
	DATA map[string]map[int64][]Data
	// Map from user to ordered list of existing timestamps.
	DATA_SERIES map[string][]int64

	// List of all metrics with basic statistics.
	METRICS map[string]*MetricType

	// Store whitelisted data.
	METRICS_DATA map[string]*MetricsData

	// Map of users.
	USERS map[string]User

	// User specs.
	SPEC Spec

	// Server spec
	STORE_TTL int64

	DATA_MUX sync.Mutex

	I StringInterner
)

func Init() {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()
	DATA = make(map[string]map[int64][]Data)
	DATA_SERIES = make(map[string][]int64)
	METRICS = make(map[string]*MetricType)
	METRICS_DATA = make(map[string]*MetricsData)
	USERS = make(map[string]User)
	//I = StringInterner{m: make(map[string]*InternedString)}
	SPEC = Spec{
		SampleInterval: SECOND_MS,
		StoreInterval:  MINUTE_MS,
		MetricsWhitelist: []string{
			"[name:audio].reports.[type:remote-inbound-rtp].jitter",
			"[name:audio].reports.[type:remote-inbound-rtp].packetsLost",
			"[name:audio].reports.[type:remote-inbound-rtp].roundTripTime",
			"[name:video].reports.[type:remote-inbound-rtp].jitter",
			"[name:video].reports.[type:remote-inbound-rtp].packetsLost",
			"[name:video].reports.[type:remote-inbound-rtp].roundTripTime",
			"[name:Misc].reports.[type:misc].slow-link-receiving",
			"[name:Misc].reports.[type:misc].slow-link-receiving-lost",
			"[name:Misc].reports.[type:misc].iceState",
		},
	}
	STORE_TTL = TEN_MINUTES_MS
}

type InternedString struct {
	s string
	c int64
}

type StringInterner struct {
	m           map[string]*InternedString
	added       int64
	deleted     int64
	totalBytes  int64
	storedBytes int64
	mux         sync.Mutex
}

func (si *StringInterner) I(s string) string {
	si.mux.Lock()
	defer si.mux.Unlock()
	if _, ok := si.m[s]; !ok {
		si.m[s] = &InternedString{s: s, c: int64(0)}
		si.added++
		si.storedBytes += int64(len(s))
	}
	si.m[s].c++
	si.totalBytes += int64(len(s))
	return si.m[s].s
}

func (si *StringInterner) DI(s string) {
	si.mux.Lock()
	defer si.mux.Unlock()
	si.m[s].c--
	si.totalBytes -= int64(len(s))
	if si.m[s].c == 0 {
		delete(si.m, s)
		si.deleted++
		si.storedBytes -= int64(len(s))
	}
}

func (si *StringInterner) Info() {
	si.mux.Lock()
	defer si.mux.Unlock()
	log.Infof("String interner. Size: %d Added: %d, Deleted: %d, Total bytes: %d, Saved bytes: %d Used bytes: %d",
		len(si.m), si.added, si.deleted, si.totalBytes, si.totalBytes-si.storedBytes, si.storedBytes)
}

func GetUser(r map[string]interface{}) (User, error) {
	if user, ok := r["user"]; !ok {
		return nil, errors.New(fmt.Sprintf("Expected 'user' object got: %+v.", r))
	} else {
		if u, ok := user.(User); !ok {
			return nil, errors.New("Expected 'user' to be an object.")
		} else {
			return u, nil
		}
	}
}

func UserId(u User) (string, error) {
	if id, ok := u["id"]; !ok {
		return "", errors.New("Expected 'id' in user object.")
	} else {
		if idString, ok := id.(string); !ok || idString == "" {
			return "", errors.New("Expected 'id' to be a non empty string.")
		} else {
			return idString, nil
		}
	}
}

func GetDatas(r map[string]interface{}) ([][]Data, error) {
	if data, ok := r["data"]; !ok {
		return nil, errors.New("Expected 'data'.")
	} else {
		if dataTyped, ok := data.([]interface{}); !ok {
			return nil, errors.New("Expected 'data' to be an array.")
		} else {
			d := [][]Data(nil)
			for _, dataElem := range dataTyped {
				if dataElemTyped, ok := dataElem.([]interface{}); !ok {
					return nil, errors.New("Expected 'data' to be an array of array.")
				} else {
					subD := []Data(nil)
					for _, dataElemOfElem := range dataElemTyped {
						if dataElemOfElemTyped, ok := dataElemOfElem.(Data); !ok {
							return nil, errors.New("Expected 'data' to be an array of array of objects.")
						} else {
							subD = append(subD, dataElemOfElemTyped)
						}
					}
					d = append(d, subD)
				}
			}
			return d, nil
		}
	}
}

func ObjKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func GetInt64(json map[string]interface{}, field string) (int64, error) {
	if value, ok := json[field]; !ok {
		return 0, errors.New(fmt.Sprintf("Expected '%s'.", field))
	} else {
		if v, ok := value.(int64); !ok {
			return 0, errors.New(fmt.Sprintf("Expected '%s' to be int64 got '%s'.", field, reflect.TypeOf(value)))
		} else {
			return v, nil
		}
	}
}

func GetTimestamp(v map[string]interface{}) (int64, error) {
	f, err := GetFloat64(v, "timestamp")
	if err != nil {
		return 0, err
	}
	return int64(f), nil
}

func GetString(json map[string]interface{}, field string) (string, error) {
	if value, ok := json[field]; !ok {
		return "", errors.New(fmt.Sprintf("Expected '%s' have: %+v", field, ObjKeys(json)))
	} else {
		if v, ok := value.(string); !ok {
			return "", errors.New(fmt.Sprintf("Expected '%s' to be string got '%s'.", field, reflect.TypeOf(value)))
		} else {
			return v, nil
		}
	}
}

func GetFloat64(json map[string]interface{}, field string) (float64, error) {
	if value, ok := json[field]; !ok {
		return 0, errors.New(fmt.Sprintf("Expected '%s'.", field))
	} else {
		if v, ok := value.(float64); !ok {
			return 0, errors.New(fmt.Sprintf("Expected '%s' to be float64 got '%s'.", field, reflect.TypeOf(value)))
		} else {
			return v, nil
		}
	}
}
