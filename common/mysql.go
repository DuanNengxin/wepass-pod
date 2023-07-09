package common

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"pod/config"
	"pod/domain/model"
)

var mysqlDB *gorm.DB

func InitMysql(migrate bool) *gorm.DB {
	var err error
	mysqlConfig := config.Config().Mysql
	mysqlDB, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		mysqlConfig.Username, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.DB, "utf8mb4")), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if migrate {
		autoMigrate()
	}
	return mysqlDB
}

func autoMigrate() {
	err := mysqlDB.AutoMigrate(
		&model.Pod{}, &model.PodEnv{}, &model.PodPort{},
	)
	if err != nil {
		return
	}
}

func CloseMysql() error {
	pool, err := mysqlDB.DB()
	if err != nil {
		panic(err)
	}
	return pool.Close()
}
