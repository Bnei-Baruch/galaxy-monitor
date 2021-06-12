package common

import log "github.com/sirupsen/logrus"

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	//log.SetLevel(log.WarnLevel)
}
