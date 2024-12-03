package main

import (
	"github.com/robfig/cron/v3"
	"log"
)

func syncData() {
	config := GetGlobalConfig()
	if config.Cron == "" {
		log.Println("不支持定时同步监听服务", config.Cron)
		return
	}
	if config.GroupAddr == "" {
		log.Println("当前集群名称为空，不支持同步操作", config.GroupAddr)
		return
	}
	log.Println("支持定时同步监听服务", config.Cron, "GroupAddr", config.GroupAddr)

	addSyncDataTimer(config)
}

func addSyncDataTimer(config Config) {
	// 函数没执行完就跳过本次函数
	c := cron.New(cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	// 添加一个任务，每 30 分钟 执行一次
	id, err := c.AddFunc(config.Cron, func() { log.Println("Every hour on the half hour") })
	if err != nil {
		log.Println("func id", id, "add func error:", err.Error())
		return
	}
	// 开始执行（每个任务会在自己的 goroutine 中执行）
	c.Start()
}
