package main

import (
	"fmt"
	"github.com/jackcvr/gpstrack/orm"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
	_ "time/tzdata"
)

const (
	LogFlags   = log.LstdFlags | log.Lshortfile | log.Lmicroseconds
	BufferSize = 1024
)

var (
	HTTPPort = os.Args[1]
	GPRSPort = os.Args[2]
	Debug    = os.Getenv("DEBUG")
	errLog   = log.New(os.Stderr, "ERROR: ", LogFlags)
)

func init() {
	loc, err := time.LoadLocation(orm.TZ)
	if err != nil {
		panic(err)
	}
	time.Local = loc
	log.SetFlags(LogFlags)
	if Debug != "" && Debug != "0" && Debug != "false" {
		orm.Client.Logger = logger.Default.LogMode(logger.Info)
	}
}

func main() {
	go StartHTTPServer(fmt.Sprintf("0.0.0.0:%s", HTTPPort))
	StartTCPServer(fmt.Sprintf("0.0.0.0:%s", GPRSPort))
}
