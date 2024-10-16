package main

import (
	"flag"
	"github.com/jackcvr/gpsmap/orm"
	"github.com/pelletier/go-toml/v2"
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
	configPath = "/etc/gpsmap/gpsmap.toml"
	errLog     = log.New(os.Stderr, "", LogFlags)
)

type HTTPConfig struct {
	Bind     string
	CertFile string
	KeyFile  string
	Username string
	Password string
}

type GPRSConfig struct {
	Bind string
}

type Config struct {
	DBFile string
	TZ     string
	Debug  bool
	HTTP   HTTPConfig
	GPRS   GPRSConfig
}

var config = Config{
	DBFile: "/var/lib/gpsmap/db.sqlite3",
	TZ:     "Europe/Vilnius",
	HTTP: HTTPConfig{
		Bind:     "0.0.0.0:12000",
		CertFile: "server.crt",
		KeyFile:  "server.key",
		Username: "admin",
		Password: "admin",
	},
	GPRS: GPRSConfig{
		Bind: "0.0.0.0:12050",
	},
}

func init() {
	log.SetFlags(LogFlags)
}

func main() {
	flag.StringVar(&configPath, "c", configPath, "Path to TOML config file")
	flag.Parse()

	if data, err := os.ReadFile(configPath); err != nil {
		log.Fatal(err)
	} else if err = toml.Unmarshal(data, &config); err != nil {
		log.Fatal(err)
	}

	if loc, err := time.LoadLocation(config.TZ); err != nil {
		panic(err)
	} else {
		time.Local = loc
	}

	db := orm.GetClient(config.DBFile)
	if config.Debug {
		db.Logger = logger.Default.LogMode(logger.Info)
	}

	go StartHTTPServer(db, config.HTTP)
	StartTCPServer(db, config.GPRS)
}
