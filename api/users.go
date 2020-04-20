package api

import (
	"database/sql"
	"math"

	log "github.com/Sirupsen/logrus"
)

type UsersResponse struct {
	Users []User `json:"users"`
}

func handleUsers(db *sql.DB) (*UsersResponse, *HttpError) {
	DATA_MUX.Lock()
	defer DATA_MUX.Unlock()

	res := UsersResponse{Users: []User{}}
	for _, user := range USERS {
		res.Users = append(res.Users, user)
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

	for userId, md := range METRICS_DATA {
		if len(md.Timestamps) > 0 {
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

	for k, v := range res.UsersData {
		log.Infof("User: %s", k)
		log.Infof("Index: %+v", v.Index)
		log.Infof("Timestamps: %+v", v.Timestamps)
		log.Infof("Data: %+v", v.Data)
		for i := range v.Stats {
			for j := range v.Stats[i] {
				s := v.Stats[i][j]
				log.Infof("Stats[%d][%d]: %+v", i, j, s)

				if math.IsNaN(s.Mean) || math.IsNaN(s.DSquared) {
					log.Infof("NaN!")
				}
				if math.IsInf(s.Mean, 1) || math.IsInf(s.DSquared, 1) {
					log.Infof("+Inf!")
				}
				if math.IsInf(s.Mean, -1) || math.IsInf(s.DSquared, -1) {
					log.Infof("-Inf!")
				}
			}
		}
	}

	return &res, nil
}
