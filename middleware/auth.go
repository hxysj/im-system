package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/models"
	"github.com/hxysj/im-system/utils"
)

func AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorization := strings.TrimSpace(ctx.GetHeader("Authorization"))

		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "请先登录", "data": nil})
			return
		}

		rawToken := strings.TrimSpace(parts[1])
		claims, err := utils.ValidateToken(ctx.Request.Context(), rawToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "登录失效", "data": nil})
			return
		}

		var count int64
		err = utils.DB.
			Model(&models.UserBasic{}).
			Where("user_id = ?", claims.UserId).
			Count(&count).Error

		if err != nil || count != 1 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": -1,
				"msg":  "账号不存在或已注销",
				"data": nil,
			})
			return
		}

		ctx.Set("current_user_id", claims.UserId)
		ctx.Set("current_token_id", claims.ID)
		ctx.Next()
	}
}

func WebSocketAuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawToken := strings.TrimSpace(ctx.Query("token"))
		if rawToken == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "请先登录"})
			return
		}

		claims, err := utils.ValidateToken(ctx.Request.Context(), rawToken)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "登录失效"})
			return
		}

		var count int64
		err = utils.DB.
			Model(&models.UserBasic{}).
			Where("user_id = ?", claims.UserId).
			Count(&count).Error

		if err != nil || count != 1 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": -1,
				"msg":  "账户不存在或已注销",
				"data": nil,
			})
			return
		}

		ctx.Set("current_user_id", claims.UserId)
		ctx.Set("current_token_id", claims.ID)
		ctx.Next()
	}
}
