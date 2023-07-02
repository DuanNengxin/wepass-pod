package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type config struct {
	Mysql   *Mysql        `yaml:"mysql"`
	Redis   *Redis        `yaml:"redis"`
	MongoDB *MongoDB      `yaml:"mongoDB"`
	Consul  *ConsulConfig `yaml:"consul"`
}

type Mysql struct {
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	DB          string `yaml:"db"`
	Charset     string `yaml:"charset"`
	MaxOpenConn int    `yaml:"max_open_conn"`
	MaxIdleConn int    `yaml:"max_idle_conn"`
}

type MongoDB struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

type ConsulConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

var c config

func ParseConfig() {
	v := viper.New()
	configFileName := fmt.Sprintf("config/config.yaml")
	v.SetConfigFile(configFileName)
	if err := v.ReadInConfig(); err != nil {
		zap.S().Panic(err)
	}
	if err := v.Unmarshal(&c); err != nil {
		zap.S().Panic(err)
	}

	// viper 动态监听文件变化
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		if err := v.ReadInConfig(); err != nil {
			zap.S().Panic(err)
		}
		if err := v.Unmarshal(&c); err != nil {
			zap.S().Panic(err)
		}
		zap.S().Infof("config %v", c)
		zap.S().Infof("重新初始化redis")
		//dao.InitRedis()
		//dao.InitMysql(DBMigrate)
	})
}

func Config() *config {
	return &c
}
