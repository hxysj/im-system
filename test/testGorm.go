package main

import (
	"fmt"

	"github.com/hxysj/im-system/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 连接数据库
	db, err := gorm.Open(mysql.Open("root:1qaz@WSX3edc@tcp(127.0.0.1:3306)/im_system?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}
	// 自动迁移数据库
	db.AutoMigrate(&models.UserBasic{})

	user := &models.UserBasic{}
	user.Name = "test"
	db.Create(user)

	fmt.Println(db.First(user, 1))

	db.Model(user).Update("Password", "1234")
	fmt.Println(db.First(user, 1))
}
