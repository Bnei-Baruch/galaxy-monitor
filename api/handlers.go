package api

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

	resp, httpErr := handleUpdate(r)
	concludeRequest(c, resp, httpErr)
}

func UsersHandler(c *gin.Context) {
	resp, httpErr := handleUsers()
	concludeRequest(c, resp, httpErr)
}

func UserDataHandler(c *gin.Context) {
	var r UserDataRequest
	if c.Bind(&r) != nil {
		return
	}
	resp, httpErr := handleUserData(r)
	concludeRequest(c, resp, httpErr)
}

func UsersDataHandler(c *gin.Context) {
	resp, httpErr := handleUsersData()
	concludeRequest(c, resp, httpErr)
}

func MetricsHandler(c *gin.Context) {
	resp, httpErr := handleMetrics()
	concludeRequest(c, resp, httpErr)
}

func SpecHandlerPost(c *gin.Context) {
	var r Spec
	if c.Bind(&r) != nil {
		return
	}
	resp, httpErr := handleSpecPost(r)
	concludeRequest(c, resp, httpErr)
}

func SpecHandlerGet(c *gin.Context) {
	resp, httpErr := handleSpecGet()
	concludeRequest(c, resp, httpErr)
}

func UserMetricsHandler(c *gin.Context) {
	var r UserDataRequest
	if c.Bind(&r) != nil {
		return
	}
	resp, httpErr := handleUserMetrics(r)
	concludeRequest(c, resp, httpErr)
}

func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
