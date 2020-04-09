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
