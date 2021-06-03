package common

import (
	"time"

	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq"

	"github.com/Bnei-Baruch/galaxy-monitor/core"
)

var (
	STORE *core.Store
)

func Init() time.Time {
	clock := time.Now()

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	//log.SetLevel(log.WarnLevel)

	STORE = &core.Store{}

	return clock
}
