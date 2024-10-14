package orm

import (
	"github.com/glebarez/sqlite"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
	"os"
)

var (
	DBName = os.Getenv("DBNAME")
	Client *gorm.DB
)

func init() {
	var err error
	Client, err = gorm.Open(sqlite.Open(DBName), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if err = Client.AutoMigrate(
		&Record{},
	); err != nil {
		panic(err)
	}
}
