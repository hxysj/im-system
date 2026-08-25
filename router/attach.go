package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/service"
)

func RegisterAttachRoutes(r *gin.RouterGroup) {
	attach := r.Group("/attach")
	{
		attach.POST("/upload", service.Upload)
	}

}
