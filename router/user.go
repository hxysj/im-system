package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/service"
)

func RegisterPublicUserRoutes(r *gin.Engine) {
	user := r.Group("/user")
	{
		// 登录注册模块
		user.POST("/register", service.CreateUser)
		user.POST("/login", service.Login)
	}
}

func RegisterUserRoutes(r *gin.RouterGroup) {
	user := r.Group("/user")
	{
		// 用户模块
		user.GET("/getUserList", service.GetUserList)
		user.GET("/deleteUser", service.DeleteUser)
		user.POST("/update", service.UpdateUser)
		user.POST("/searchUser", service.SearchUser)
		user.POST("/changePassword", service.UpdatePassword)
	}
}
