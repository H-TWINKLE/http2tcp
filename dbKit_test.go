package main

import (
	"log"
	"testing"
)

func TestDBInsert(t *testing.T) {
	db := GetDB()
	if db == nil {
		t.Error("db is nil")
		return
	}

	proxyConfig := ProxyConfig{
		ServerAddr: "http://127.0.0.1:8080",
		GroupAddr:  "127.0.0.1",
		Token:      "longlongauthtoken",
		TargetHost: "127.0.0.1",
		TargetPort: 3308,
		ListenAddr: "127.0.0.1:13308",
		Mode:       "tcp"}
	result := db.Create(&proxyConfig)

	if result.Error != nil {
		log.Println("插入数据失败:", result.Error)
	} else {
		log.Println("新代理信息 ID:", proxyConfig.ID)
	}
}

func TestDBQuery(t *testing.T) {
	db := GetDB()
	if db == nil {
		t.Error("db is nil")
		return
	}

	var proxyConfig []ProxyConfig
	result := db.Find(&proxyConfig)

	if result.Error != nil {
		log.Println("获取数据失败:", result.Error)
	} else {
		log.Println("获取数据:", proxyConfig)
	}
}

func TestDBUpdate(t *testing.T) {
	db := GetDB()
	if db == nil {
		t.Error("db is nil")
		return
	}

	var proxyConfig ProxyConfig
	result := db.Find(&proxyConfig)

	if result.Error != nil {
		log.Println("获取数据失败:", result.Error)
	} else {
		log.Println("获取数据:", proxyConfig)
		proxyConfig.Status = 1
		db.Save(&proxyConfig)
	}

}

func TestDBDelete(t *testing.T) {
	db := GetDB()
	if db == nil {
		t.Error("db is nil")
		return
	}

	var proxyConfig ProxyConfig
	result := db.Find(&proxyConfig)

	if result.Error != nil {
		log.Println("获取数据失败:", result.Error)
	} else {
		log.Println("获取数据:", proxyConfig)
		// 根据主键删除数据（逻辑删除）
		db.Delete(&ProxyConfig{}, 2)
		// 根据主键删除数据（永久删除）
		// db.Unscoped().Delete(&ProxyConfig{}, 2)
	}

}
