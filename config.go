package main

import (
	"fmt"
	"github.com/jinzhu/configor"
	"os"
	"path"
)

type Config struct {
	APPName   string `default:"http2tcp-app"`
	Cron      string
	GroupAddr string
	Notice    struct {
		Server []string
	}
	IsInit bool
}

var globalConfig Config

func IsInitBy(config Config) bool {
	return config.IsInit
}

func GetCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err) // 错误处理
	}
	return dir
}

func InitGlobalConfig() Config {
	curr := GetCurrentDirectory()
	dataPath := path.Join(curr, "config.yml")

	var conf = Config{}
	err := configor.Load(&conf, dataPath)
	if err != nil {
		fmt.Println("config.yml (", dataPath, ") parse error:", err)
		return Config{}
	}
	globalConfig = conf
	return globalConfig
}

// GetGlobalConfig 获取全局配置的函数
func GetGlobalConfig() Config {
	if !IsInitBy(globalConfig) {
		InitGlobalConfig()
	}
	return globalConfig
}
