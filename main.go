package main

import (
	"log"

	"github.com/hxysj/im-system/models"
	"github.com/hxysj/im-system/router"
	"github.com/hxysj/im-system/utils"
)

func main() {
	utils.InitApp()
	utils.InitRedis()

	if err := utils.InitMysql(); err != nil {
		log.Fatal(err)
	}

	// 启动时自动创建或更新用户表结构
	if err := utils.DB.AutoMigrate(&models.UserBasic{}); err != nil {
		log.Fatal(err)
	}

	r := router.Router()
	r.Run()

}
