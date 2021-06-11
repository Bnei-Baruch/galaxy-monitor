package common

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func Init() time.Time {
	clock := time.Now()

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	//log.SetLevel(log.WarnLevel)

	return clock
}
