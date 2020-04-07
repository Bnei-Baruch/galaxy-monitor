package api

import (
	"compress/flate"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type Data = map[string]interface{}
type User = map[string]interface{}

const (
	STORE_TTL int64 = 5 * 60 * 1000 // 5 minutes back, in ms.
)

var (
	// Map from user to timestamp to data.
	DATA map[string]map[int64][]Data
	// Map from user to ordered list of existing timestamps
	DATA_SERIES map[string][]int64

	// Map of users.
	USERS map[string]User

	DATA_MUX sync.Mutex

	I StringInterner
)

type StringInterner struct {
	m   map[string]string
	mux sync.Mutex
}

func (si *StringInterner) I(s string) string {
	si.mux.Lock()
	defer si.mux.Unlock()
	if _, ok := si.m[s]; !ok {
		si.m[s] = s
	}
	return si.m[s]
}

func Init() {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()
	DATA = make(map[string]map[int64][]Data)
	DATA_SERIES = make(map[string][]int64)
	USERS = make(map[string]User)
	I = StringInterner{m: make(map[string]string)}
}

func InternJsonValue(v interface{}) interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return InternJsonMap(m)
	} else if a, ok := v.([]interface{}); ok {
		return InternJsonArr(a)
	} else if s, ok := v.(string); ok {
		return I.I(s)
	} else {
		return v
	}
}

func InternJsonMap(json map[string]interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range json {
		m[I.I(k)] = InternJsonValue(v)
	}
	return m
}

func InternJsonArr(json []interface{}) []interface{} {
	a := make([]interface{}, 0)
	for i := range json {
		a = append(a, InternJsonValue(json[i]))
	}
	return a
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

func GetDatas(r map[string]interface{}) ([]Data, error) {
	if data, ok := r["data"]; !ok {
		return nil, errors.New("Expected 'data'.")
	} else {
		if dataTyped, ok := data.([]interface{}); !ok {
			return nil, errors.New("Expected 'data' to be an array.")
		} else {
			d := []Data(nil)
			for _, dataElem := range dataTyped {
				if dataElemTyped, ok := dataElem.([]interface{}); !ok {
					return nil, errors.New("Expected 'data' to be an array of array.")
				} else {
					for _, dataElemOfElem := range dataElemTyped {
						if dataElemOfElemTyped, ok := dataElemOfElem.(Data); !ok {
							return nil, errors.New("Expected 'data' to be an array of array of objects.")
						} else {
							d = append(d, dataElemOfElemTyped)
						}
					}
				}
			}
			return d, nil
		}
	}
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

// Responds with JSON of given response or aborts the request with the given error.
func concludeRequest(c *gin.Context, resp interface{}, err *HttpError) {
	if err == nil {
		c.JSON(http.StatusOK, resp)
	} else {
		err.Abort(c)
	}
}

type UpdateResponse struct{}

func UpdateHandler(c *gin.Context) {
	var reader io.Reader
	switch c.Request.Header.Get("Content-Encoding") {
	case "gzip":
		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			log.Infof("error with gzip reader: %s", err.Error())
		}
		defer gz.Close()
		reader = gz
	case "deflate":
		def := flate.NewReader(c.Request.Body)
		defer def.Close()
		reader = def
	default:
		// just use the default reader
		reader = c.Request.Body
	}
	// NOTE: assuming json here
	// you may want to check r.Header.Get("Content-Type")
	decoder := json.NewDecoder(reader)
	var v map[string]interface{}
	if err := decoder.Decode(&v); err != nil {
		log.Infof("error decoding: %s", err.Error())
		concludeRequest(c, UpdateResponse{}, NewBadRequestError(err))
		return
	}

	internR := InternJsonMap(v)

	resp, httpErr := handleUpdate(c.MustGet("MDB_DB").(*sql.DB), internR)
	concludeRequest(c, resp, httpErr)
}

func handleUpdate(db *sql.DB, r map[string]interface{}) (*UpdateResponse, *HttpError) {
	user, err := GetUser(r)
	if err != nil {
		return &UpdateResponse{}, NewBadRequestError(err)
	}
	userId, err := UserId(user)
	if err != nil {
		return &UpdateResponse{}, NewBadRequestError(err)
	}
	datas, err := GetDatas(r)
	if err != nil {
		return &UpdateResponse{}, NewBadRequestError(err)
	}
	maxTimestamp := int64(0)
	for _, data := range datas {
		floatTimestamp, err := GetFloat64(data, "timestamp")
		if err != nil {
			return &UpdateResponse{}, NewBadRequestError(err)
		}
		timestamp := int64(floatTimestamp)
		AddData(user, userId, timestamp, data)
		if timestamp > maxTimestamp {
			maxTimestamp = timestamp
		}
	}
	ClearOld(userId, maxTimestamp)

	return &UpdateResponse{}, nil
}
