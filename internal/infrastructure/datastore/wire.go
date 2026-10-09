package datastore

import (
	"github.com/google/wire"
	"gorm.io/gorm"
)

// WireSet holds the dependencies for the datastore.
var WireSet = wire.NewSet(
	NewDB,
	ProvideGormDB,
	NewBinder,
	NewTransactor,
	NewUserReader,
	NewUserWriter,
)

// ProvideGormDB provides the underlying gorm.DB.
func ProvideGormDB(db *DB) *gorm.DB {
	return db.DB
}
