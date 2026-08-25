package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/service"
)

func RegisterRelationRequest(r *gin.RouterGroup) {
	relation := r.Group("relation")
	{
		relation.GET("/getRelationList", service.GetRelationRequestList)
		relation.POST("/toggleRelationRequest", service.ToggleRelationRequestStatus)
	}
}
