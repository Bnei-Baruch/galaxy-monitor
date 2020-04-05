package common

import (
	"database/sql"
	"time"

	"github.com/Bnei-Baruch/sqlboiler/boil"
	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"

	"github.com/bbfsdev/galaxy-monitor/core"
	"github.com/bbfsdev/galaxy-monitor/utils"
)

var (
	DB    *sql.DB
	STORE *core.Store
)

func Init() time.Time {
	return InitWithDefault(nil)
}

func InitWithDefault(defaultDb *sql.DB) time.Time {
	var err error
	clock := time.Now()

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	//log.SetLevel(log.WarnLevel)

	if defaultDb != nil {
		DB = defaultDb
	} else {
		log.Info("Setting up connection to DB")
		if connectionString := viper.GetString("db.url"); connectionString == "" {
			log.Warn("Connection string is empty, skipping connection to DB.")
		} else {
			DB, err = sql.Open("postgres", connectionString)
			utils.Must(err)
			utils.Must(DB.Ping())
		}
	}
	boil.SetDB(DB)
	boil.DebugMode = viper.GetString("server.boiler-mode") == "debug"

	STORE = &core.Store{}

	return clock
}

func Shutdown() {
	utils.Must(DB.Close())
}
