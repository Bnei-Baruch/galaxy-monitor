package api

import (
	"database/sql"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

// Responds with JSON of given response or aborts the request with the given error.
func concludeRequest(c *gin.Context, resp interface{}, err *HttpError) {
	if err == nil {
		c.JSON(http.StatusOK, resp)
	} else {
		err.Abort(c)
	}
}

type UpdateRequest struct {
}

type UpdateResponse struct {
}

func UpdateHandler(c *gin.Context) {
	r := UpdateRequest{}
	if c.Bind(&r) != nil {
		return
	}

	resp, err := handleUpdate(c.MustGet("MDB_DB").(*sql.DB), r)
	concludeRequest(c, resp, err)
}

func handleUpdate(db *sql.DB, r UpdateRequest) (*UpdateResponse, *HttpError) {
	log.Infof("r: %+v", r)
	return &UpdateResponse{}, nil
}
