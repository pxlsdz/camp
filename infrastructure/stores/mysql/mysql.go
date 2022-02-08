package mysql

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func Init() error {
	var err error

	var cfgs = []string{"username", "password", "addr", "database"}
	var cfgVals = make([]interface{}, 0)
	for _, cfg := range cfgs {
		cfgVals = append(cfgVals, viper.GetString("mysql."+cfg))
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfgVals...)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	return err
}

func GetDb() *gorm.DB {
	return db
}
