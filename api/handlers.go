package api

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func UpdateHandler(c *gin.Context) {
	r, err := readGzipJson(c.Request)
	if err != nil {
		log.Errorf("Error reading request: %s", err.Error())
		concludeRequest(c, nil, NewBadRequestError(err))
		return
	}

	resp, httpErr := handleUpdate(r)
	concludeRequest(c, resp, httpErr)
}

func UsersHandler(c *gin.Context) {
	resp, err := handleUsers()
	concludeRequest(c, resp, err)
}

func UserDataHandler(c *gin.Context) {
	var r UserDataRequest
	if c.Bind(&r) != nil {
		return
	}
	resp, err := handleUserData(r)
	concludeRequest(c, resp, err)
}

func UsersDataHandler(c *gin.Context) {
	resp, err := handleUsersData()
	concludeRequest(c, resp, err)
}

func MetricsHandler(c *gin.Context) {
	resp, err := handleMetrics()
	concludeRequest(c, resp, err)
}

func SpecHandlerPost(c *gin.Context) {
	var r Spec
	if c.Bind(&r) != nil {
		return
	}
	concludeRequest(c, gin.H{"status": "ok"}, handleSpecPost(r))
}

func SpecHandlerGet(c *gin.Context) {
	resp, err := handleSpecGet()
	concludeRequest(c, resp, err)
}

func UserMetricsHandler(c *gin.Context) {
	var r UserDataRequest
	if c.Bind(&r) != nil {
		return
	}
	resp, err := handleUserMetrics(r)
	concludeRequest(c, resp, err)
}

func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Responds with JSON of given response or aborts the request with the given error.
func concludeRequest(c *gin.Context, resp interface{}, err *HttpError) {
	if err == nil {
		c.JSON(http.StatusOK, resp)
	} else {
		err.Abort(c)
	}
}

// Read and decode JSON request body (maybe gzip)
func readGzipJson(r *http.Request) (map[string]interface{}, error) {
	var reader io.Reader
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, errors.Wrap(err, "gzip.NewReader")
		}
		defer gz.Close()
		reader = gz
	default:
		// Just use the default reader.
		reader = r.Body
	}

	// NOTE: Assuming json here, should check "Content-Type"
	var v map[string]interface{}
	if err := json.NewDecoder(reader).Decode(&v); err != nil {
		return nil, errors.Wrap(err, "json.Decode")
	}

	return v, nil
}
