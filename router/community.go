package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/service"
)

func RegisterCommunityRoutes(r *gin.Engine){
	community := r.Group("/community")
	{
		community.POST("/createCommunity",service.CreateCommunity)
		community.POST("/getCommunityList",service.LoadCommunity)
		community.POST("/joinCommunity",service.CreateCommunityRelation)
	}
}