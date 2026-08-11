package main

import (
	"github.com/hxysj/im-system/router"
	"github.com/hxysj/im-system/utils"
)

func main() {
	utils.InitApp()
	utils.InitMysql()

	

	r := router.Router()
	r.Run()

}