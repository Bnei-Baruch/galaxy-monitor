package api

import (
	"database/sql"
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
