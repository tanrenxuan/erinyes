package models

import (
	"erinyes/conf"
	"erinyes/logs"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
	"time"
)

var (
	_db *gorm.DB
	Mu  sync.Mutex
)

func Init() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.Config.Mysql.Username,
		conf.Config.Mysql.Password,
		conf.Config.Mysql.Host,
		conf.Config.Mysql.Port,
		conf.Config.Mysql.DBName)
	var err error
	_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败，error=" + err.Error())
	}
	_db.Logger.LogMode(logger.Silent)
	sqlDB, err := _db.DB()
	if err != nil {
		panic("获取数据库sqlDB失败，error=" + err.Error())
	}
	sqlDB.SetMaxIdleConns(conf.Config.Mysql.MaxIdleConns)
	sqlDB.SetMaxOpenConns(conf.Config.Mysql.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		panic("无法ping数据库: " + err.Error())
	}
	logs.Logger.Info("成功连接到数据库")
}

func GetMysqlDB() *gorm.DB {
	return _db
}
