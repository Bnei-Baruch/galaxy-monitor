package api

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

type EmptyResponse struct{}

// Responds with JSON of given response or aborts the request with the given error.
func concludeRequest(c *gin.Context, resp interface{}, err *HttpError) {
	if err == nil {
		c.JSON(http.StatusOK, resp)
	} else {
		err.Abort(c)
	}
}

func readGzipJson(r *http.Request) (map[string]interface{}, error) {
	var reader io.Reader
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			log.Infof("error with gzip reader: %s", err.Error())
		}
		defer gz.Close()
		reader = gz
	default:
		// Just use the default reader.
		reader = r.Body
	}
	// NOTE: Assuming json here, should check "Content-Type"
	decoder := json.NewDecoder(reader)

	var v map[string]interface{}
	if err := decoder.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

func UpdateHandler(c *gin.Context) {
	r, err := readGzipJson(c.Request)
	if err != nil {
		log.Infof("Error reading request: %s", err.Error())
		concludeRequest(c, EmptyResponse{}, NewBadRequestError(err))
		return
	}

	internR := InternJsonMap(r)

	resp, httpErr := handleUpdate(c.MustGet("MDB_DB").(*sql.DB), internR)
	concludeRequest(c, resp, httpErr)
}

func UsersHandler(c *gin.Context) {
	resp, httpErr := handleUsers(c.MustGet("MDB_DB").(*sql.DB))
	concludeRequest(c, resp, httpErr)
}

func UserDataHandler(c *gin.Context) {
	var r UserDataRequest
	if c.Bind(&r) != nil {
		return
	}
	resp, httpErr := handleUserData(c.MustGet("MDB_DB").(*sql.DB), r)
	concludeRequest(c, resp, httpErr)
}

func MetricsHandler(c *gin.Context) {
	resp, httpErr := handleMetrics(c.MustGet("MDB_DB").(*sql.DB))
	concludeRequest(c, resp, httpErr)
}
