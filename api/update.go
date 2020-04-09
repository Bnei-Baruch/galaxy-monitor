package api

import (
	"database/sql"
	"sort"

	log "github.com/Sirupsen/logrus"
)

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
	}
	ClearOld(userId, maxTimestamp)

	return &EmptyResponse{}, nil
}
