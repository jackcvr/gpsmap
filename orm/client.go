package orm

import (
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"time"
)

const (
	DSN = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=%s"
	TZ  = "Europe/Vilnius"
)

var (
	PGHost     = os.Getenv("PG_HOST")
	PGUser     = os.Getenv("PG_USER")
	PGDatabase = os.Getenv("PG_DATABASE")
	PGPassword = os.Getenv("PG_PASSWORD")
	PGPort     = os.Getenv("PG_PORT")
	Client     *gorm.DB
)

func init() {
	if PGPort == "" {
		PGPort = "5432"
	}
	var err error
	dsn := fmt.Sprintf(DSN, PGHost, PGPort, PGUser, PGPassword, PGDatabase, TZ)
	for _ = range 4 {
		Client, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		} else {
			time.Sleep(time.Second)
		}
	}
	if err != nil {
		panic(err)
	}
	if err = Client.AutoMigrate(
		&Record{},
	); err != nil {
		panic(err)
	}
}
