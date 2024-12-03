package main

import (
	"github.com/glebarez/sqlite" // Sqlite driver based on CGO
	"gorm.io/gorm"
	"log"
	"path"
)

var DB *gorm.DB

func initDB() {
	curr := GetCurrentDirectory()
	dbPath := path.Join(curr, "http2tcp.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalln("init DB error:", err)
	}
	// 自动迁移
	err = db.AutoMigrate(&ProxyConfig{})
	if err != nil {
		log.Fatalln("auto migrate DB error:", err)
	}
	DB = db
}

func GetDB() *gorm.DB {
	if DB == nil {
		initDB()
	}
	return DB
}

type ProxyConfig struct {
	gorm.Model
	ID         uint `gorm:"primarykey;AUTO_INCREMENT"`
	GroupAddr  string
	ServerAddr string
	Token      string
	TargetHost string
	TargetPort int
	ListenAddr string
	Mode       string
	Status     int
}
