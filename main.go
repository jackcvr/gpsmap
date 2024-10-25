package main

import (
	"flag"
	"github.com/jackcvr/gpsmap/orm"
	"github.com/pelletier/go-toml/v2"
	"log"
	"os"
	"time"
	_ "time/tzdata"
)

const LogFlags = log.LstdFlags | log.Lshortfile | log.Lmicroseconds

var (
	configPath = "/etc/gpsmap/gpsmap.toml"
	errLog     = log.New(os.Stderr, "", LogFlags)
)

type Config struct {
	DBFile   string
	KeepDays uint
	TZ       string
	Debug    bool
	HTTP     HTTPConfig
	GPRS     GPRSConfig
	TGBot    TGBotConfig
}

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

type TGBotConfig struct {
	Token   string
	Timeout int
}

var config = Config{
	DBFile:   "/var/lib/gpsmap/db.sqlite3",
	KeepDays: 30,
	TZ:       "Europe/Vilnius",
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
	TGBot: TGBotConfig{
		Token:   "",
		Timeout: 0,
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

	db := orm.GetClient(config.DBFile, config.Debug)
	go RunPeriodic(time.Hour, func() {
		t := time.Now().Add(-time.Duration(config.KeepDays) * 24 * time.Hour)
		if err := db.Where("created_at < ?", t).Delete(&orm.Record{}).Error; err != nil {
			panic(err)
		}
		if err := db.Exec("PRAGMA optimize;").Error; err != nil {
			panic(err)
		}
		log.Printf("database %s optimized", config.DBFile)
	})
	pubsub := NewPubSub()
	go ServeHTTP(config.HTTP, db, pubsub, config.Debug)
	var bot *TGBot
	if config.TGBot.Token != "" {
		bot = StartTGBot(config.TGBot, db, config.Debug)
	}
	ServeTCP(config, db, pubsub, bot)
}
