package main

import (
	"flag"
	"github.com/jackcvr/gpstrack/orm"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
	_ "time/tzdata"
)

const (
	TZ         = "Europe/Vilnius"
	LogFlags   = log.LstdFlags | log.Lshortfile | log.Lmicroseconds
	BufferSize = 1024
)

var (
	HTTPAddr string
	GPRSAddr string
	Debug    = os.Getenv("DEBUG")
	errLog   = log.New(os.Stderr, "ERROR: ", LogFlags)
)

func init() {
	loc, err := time.LoadLocation(TZ)
	if err != nil {
		panic(err)
	}
	time.Local = loc
	log.SetFlags(LogFlags)
	if Debug != "" && Debug != "0" && Debug != "false" {
		orm.Client.Logger = logger.Default.LogMode(logger.Info)
	}

	flag.StringVar(&HTTPAddr, "http", "0.0.0.0:12000", "HTTPS address to start web interface on")
	flag.StringVar(&GPRSAddr, "gprs", "0.0.0.0:12050", "TCP address to start GPRS receiver on")
}

func main() {
	flag.Parse()
	go StartHTTPServer(HTTPAddr)
	StartTCPServer(GPRSAddr)
}
