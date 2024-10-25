package orm

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const SQLiteTuning = `PRAGMA journal_mode = WAL;
PRAGMA synchronous = normal;
PRAGMA journal_size_limit = 6144000;
PRAGMA temp_store = memory;
PRAGMA mmap_size = 30000000000;
PRAGMA optimize = 0x10002;
VACUUM;`

var clients = map[string]*gorm.DB{}

func GetClient(dsn string, debug bool) *gorm.DB {
	client, ok := clients[dsn]
	if ok {
		return client
	}
	var err error
	var _logger logger.Interface
	if debug {
		_logger = logger.Default.LogMode(logger.Info)
	}
	client, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: _logger,
	})
	if err != nil {
		panic(err)
	}
	if err = client.Exec(SQLiteTuning).Error; err != nil {
		panic(err)
	}
	if err = client.AutoMigrate(
		&Record{},
		&TGChat{},
	); err != nil {
		panic(err)
	}
	clients[dsn] = client
	return client
}
