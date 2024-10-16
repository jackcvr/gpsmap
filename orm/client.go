package orm

import (
	"github.com/glebarez/sqlite"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
)

var clients = map[string]*gorm.DB{}

func GetClient(filename string) *gorm.DB {
	client, ok := clients[filename]
	if ok {
		return client
	}
	var err error
	client, err = gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if err = client.AutoMigrate(
		&Record{},
	); err != nil {
		panic(err)
	}
	clients[filename] = client
	return client
}
