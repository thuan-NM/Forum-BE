package infrastructure

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectMySQL(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
