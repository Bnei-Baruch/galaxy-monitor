package api

import (
	"database/sql"
	"time"
)

type UsersResponse struct {
	Users []User `json:"users"`
}

func handleUsers(db *sql.DB) (*UsersResponse, *HttpError) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()

	res := UsersResponse{Users: []User{}}
	now := time.Now()

	for userId, user := range USERS {
		if md, ok := METRICS_DATA[userId]; ok && len(md.Timestamps) > 0 && now.Sub(time.Unix(md.Timestamps[len(md.Timestamps)-1]/1000, 0)).Minutes() < 2.0 {
			res.Users = append(res.Users, user)
		}
	}
	return &res, nil
}

type UserDataRequest struct {
	UserId string `json:"user_id"`
}

type UserDataResponse struct {
	Data [][]Data `json:"data"`
}

func handleUserData(db *sql.DB, r UserDataRequest) (*UserDataResponse, *HttpError) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()

	res := UserDataResponse{Data: [][]Data{}}
	if timestamps, ok := DATA_SERIES[r.UserId]; ok {
		for _, timestamp := range timestamps {
			res.Data = append(res.Data, DATA[r.UserId][timestamp])
		}
	}
	return &res, nil
}

type UserMetricsResponse struct {
	Metrics MetricsData `json:"metrics"`
}

func handleUserMetrics(db *sql.DB, r UserDataRequest) (*UserMetricsResponse, *HttpError) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()

	res := UserMetricsResponse{}
	if md, ok := METRICS_DATA[r.UserId]; ok {
		res.Metrics = *md
	}
	return &res, nil
}

type UsersDataResponse struct {
	UsersData map[string]MetricsData `json:"users_data"`
}

func handleUsersData(db *sql.DB) (*UsersDataResponse, *HttpError) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()

	res := UsersDataResponse{UsersData: make(map[string]MetricsData)}
	now := time.Now()

	for userId, md := range METRICS_DATA {
		if len(md.Timestamps) > 0 && now.Sub(time.Unix(md.Timestamps[len(md.Timestamps)-1]/1000, 0)).Minutes() < 2.0 {
			sendData := [][]interface{}(nil)
			for i := range md.Data {
				sendData = append(sendData, []interface{}{md.Data[i][len(md.Data[i])-1]})
			}
			res.UsersData[userId] = MetricsData{
				Index:      md.Index,
				Timestamps: md.Timestamps[len(md.Timestamps)-1:],
				Data:       sendData,
				Stats:      md.Stats,
			}
		}
	}

	return &res, nil
}
