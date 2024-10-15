package main

import (
	"flag"
	"github.com/jackcvr/gpsmap/orm"
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
	HTTPAddr string
	GPRSAddr string
	TZ       = os.Getenv("TZ")
	Debug    = os.Getenv("DEBUG")
	errLog   = log.New(os.Stderr, "ERROR: ", LogFlags)
)

func init() {
	if TZ == "" {
		TZ = "Europe/Vilnius"
	}
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
