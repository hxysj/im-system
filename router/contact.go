package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/service"
)

func RegisterContactRoutes(r *gin.Engine) {
	contact := r.Group("/contact")
	{
		contact.POST("/getFriendsById", service.SearchFriends)
		contact.POST("/addFriend",service.AddFriend)
	}
}