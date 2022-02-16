package mysql

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var db *gorm.DB

//func Init() error {
//	var err error
//
//	var cfgs = []string{"username", "password", "addr", "database"}
//	var cfgVals = make([]interface{}, 0)
//	for _, cfg := range cfgs {
//		cfgVals = append(cfgVals, viper.GetString("mysql."+cfg))
//	}
//	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfgVals...)
//
//	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
//	//Logger: logger.Default.LogMode(logger.Info),
//
//	return err
//}

func Init() error {
	var err error

	var cfgs = []string{"username", "password", "addr", "database"}
	var cfgVals = make([]interface{}, 0)
	for _, cfg := range cfgs {
		cfgVals = append(cfgVals, viper.GetString("mysql."+cfg))
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfgVals...)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	//Logger: logger.Default.LogMode(logger.Info),
	sqlDB, err := db.DB()

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(20)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(200)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	return err
}

func GetDb() *gorm.DB {
	return db
}
