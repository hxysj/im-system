package utils

import (
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitApp() {
	// fmt.Println("init app")
	// 设置加载的配置文件的名称
	viper.SetConfigName("app")
	// 设置加载的配置文件的位置
	viper.AddConfigPath("config")
	// 加载配置文件的内容
	err := viper.ReadInConfig()
	if err !=nil {
		fmt.Println(err)
	}

	fmt.Println("config mysql:", viper.Get("mysql"))
}

func InitMysql() {
	// fmt.Println("init mysql")
	DB,_ = gorm.Open(mysql.Open(viper.GetString("mysql.dns")),&gorm.Config{})

	// user := models.UserBasic{}
	// res := DB.Find(&user)

	// fmt.Println(res)
}