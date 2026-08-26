package main

import (
	"log"

	"github.com/hxysj/im-system/models"
	"github.com/hxysj/im-system/router"
	"github.com/hxysj/im-system/utils"
	"github.com/spf13/viper"
)

func main() {
	utils.InitApp()
	utils.InitRedis()

	if err := utils.InitMysql(); err != nil {
		log.Fatal(err)
	}

	if err := utils.InitIdGenerator(viper.GetInt64("id_generator.node_id")); err != nil {
		log.Fatal(err)
	}

	// 启动时自动创建或更新用户表结构
	if err := utils.DB.AutoMigrate(&models.UserBasic{}); err != nil {
		log.Fatal(err)
	}
	// 自动创建或更新消息表结构,群里表结构，用户关系表结构
	utils.DB.AutoMigrate(&models.Message{})
	utils.DB.AutoMigrate(&models.Contact{})
	utils.DB.AutoMigrate(&models.Community{})
	utils.DB.AutoMigrate(&models.RelationRequest{})

	r := router.Router()
	r.Run()

}
