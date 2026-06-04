package db

import (
	"syntheticvision/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open opens (creating if needed) the SQLite database at path using the
// pure-Go glebarez driver, enables foreign keys and a busy timeout, then
// auto-migrates the application schema.
func Open(path string) (*gorm.DB, error) {
	dsn := path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	// SQLite is single-writer. Pin the pool to one connection so credit-mutating
	// transactions serialize instead of racing into SQLITE_BUSY / SQLITE_BUSY_SNAPSHOT
	// errors (busy_timeout does not cover snapshot conflicts), which would surface as
	// spurious HTTP 500s under concurrent generation requests.
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	if err := gdb.AutoMigrate(
		&models.User{},
		&models.Generation{},
		&models.CreditTransaction{},
	); err != nil {
		return nil, err
	}
	return gdb, nil
}
