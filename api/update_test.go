package api

import (
	"encoding/json"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type APISuite struct {
	suite.Suite
}

func (suite *APISuite) SetupSuite() {
	Init()
}

func (suite *APISuite) TearDownSuite() {
}

func TestAPI(t *testing.T) {
	suite.Run(t, new(APISuite))
}

type ESLogAdapter struct{ *testing.T }

func (s ESLogAdapter) Printf(format string, v ...interface{}) { s.Logf(format, v...) }

func (suite *APISuite) J(jsonString string) map[string]interface{} {
	r := suite.Require()

	j := make(map[string]interface{})
	r.Nil(json.Unmarshal([]byte(jsonString), &j))

	return j
}

func (suite *APISuite) TestGetters() {
	r := suite.Require()

	j := suite.J(`{
		"user": {"id": "1"},
		"data": [
			[
				{
					"name": "audio",
					"reports": [
						{
							"id": "one",
							"type": "one type",
							"some-value": 123,
							"field": null,
							"another": ["1", "2", "3"]
						},
						{
							"id": "two",
							"type": "one type",
							"some-value": 123,
							"field": null,
							"another": ["1", "2", "3"]
						}
					],
					"timestamp": 12345
				}
			]
		]}`)

	user, err := GetUser(j)
	r.Nil(err)
	userId, err := UserId(user)
	r.Nil(err)
	r.Equal("1", userId)

	datas, err := GetDatas(j)
	log.Infof("data: %+v", datas[0])
	r.Equal(1, len(datas))

	for _, data := range datas {
		floatTimestamp, err := GetFloat64(data[0], "timestamp")
		r.Nil(err)
		r.Equal(float64(12345), floatTimestamp)
	}
}
