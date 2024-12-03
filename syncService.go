package main

import (
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

type Ret[T any] struct {
	msg     string
	code    int
	success bool
	data    T
}

type GroupInfo struct {
	groupAddr   string
	proxyConfig []ProxyConfig
}

func replyProxyInfo(w http.ResponseWriter, r *http.Request) {
	var data GroupInfo
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Println("get data error ", err.Error())
		http.Error(w, `get failed`+err.Error(), http.StatusBadRequest)
		return
	}

	db := GetDB()
	if db == nil {
		http.Error(w, "db is nil", http.StatusInternalServerError)
		log.Fatal("db is nil")
		return
	}

	SaveProxyConfig(data, db)

	info, _ := GetProxyConfig(db)
	if info.groupAddr == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	jsonBytes, _ := json.Marshal(info)
	w.Write(jsonBytes)
	w.WriteHeader(http.StatusOK)
}

func GetProxyConfig(db *gorm.DB) (GroupInfo, error) {
	config := GetGlobalConfig()
	if config.GroupAddr == "" {
		log.Println("集群名称为空，不支持同步", config.GroupAddr)
		return GroupInfo{}, nil
	}

	var proxyConfig []ProxyConfig
	result := db.Where("groupAddr=?", config.GroupAddr).Find(&proxyConfig)
	if result.Error != nil {
		log.Println("query data from db failed:", result.Error)
		return GroupInfo{}, result.Error
	}

	if len(proxyConfig) == 0 {
		return GroupInfo{}, errors.New("len is null")
	}

	info := GroupInfo{groupAddr: config.GroupAddr, proxyConfig: proxyConfig}
	return info, nil
}

func SaveProxyConfig(data GroupInfo, db *gorm.DB) {
	// 保存别人提交的信息
	if data.groupAddr != "" && len(data.proxyConfig) > 0 {
		for _, config := range data.proxyConfig {
			// 重置主键，方便同步数据
			config.ID = 0
		}
		// 更新当前的配置信息
		save := db.Save(&data.proxyConfig)

		if save.Error != nil {
			log.Println("save data from db ", data.groupAddr, " failed:", save.Error)
		}
	}
}

// 提交当前代理的运行信息
func postProxyInfo(url_ string, body []ProxyConfig) {
	config := GetGlobalConfig()
	if config.GroupAddr == "" {
		log.Println("集群名称为空，不支持同步", config.GroupAddr)
		return
	}

	client := resty.New()
	client.SetTimeout(3 * time.Second)

	var ret Ret[GroupInfo]
	resp, err := client.R().
		SetBody(GroupInfo{groupAddr: config.GroupAddr, proxyConfig: body}).
		SetResult(&ret). // or SetResult(AuthSuccess{}).
		Post(url_)

	if err != nil {
		log.Fatal(err)
		return
	}

	code := resp.StatusCode()
	if code != 200 {
		log.Fatal("resp code is :", code)
		return
	}

	var proxyInfo = ret.data

	if proxyInfo.groupAddr == "" {
		log.Fatal("groupAddr is empty")
		return
	}

	if len(proxyInfo.proxyConfig) == 0 {
		log.Fatal("proxyConfig is empty")
		return
	}

	db := GetDB()
	if db == nil {
		log.Fatal("db is nil")
		return
	}
	log.Println("prepare to delete data from ", proxyInfo.groupAddr)

	// 先删除所有的配置信息
	tx := db.Where("groupAddr=?", proxyInfo.groupAddr).Delete(&ProxyConfig{})
	if tx.Error != nil {
		log.Println("delete data from db ", proxyInfo.groupAddr, " failed:", tx.Error)
	}

	for _, config := range proxyInfo.proxyConfig {
		// 重置主键，方便同步数据
		config.ID = 0
	}
	// 更新当前的配置信息
	save := db.Save(&proxyInfo.proxyConfig)

	if save.Error != nil {
		log.Println("save data from db ", proxyInfo.groupAddr, " failed:", save.Error)
	}
}
