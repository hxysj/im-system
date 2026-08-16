package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hxysj/im-system/models"
	"github.com/hxysj/im-system/utils"
)

// 防止跨域站点伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin:func(r *http.Request) bool{
		return true
	},
}

func SendMsg(c *gin.Context){
	ws,err := upGrade.Upgrade(c.Writer,c.Request,nil)
	if err != nil{
		fmt.Println(err)
		return
	}

	defer func(ws *websocket.Conn){
		err = ws.Close()
		if err != nil{
			fmt.Println(err)
		}
	}(ws)

	MsgHandler(ws,c)
}

func MsgHandler(ws *websocket.Conn,c *gin.Context){
	for {
		msg,err := utils.Subscribe(c,utils.PublishKey)
		if err != nil{
			fmt.Println(err)
		}
		fmt.Println("发送消息 ",msg)
		tm := time.Now().Format("2006-01-02 15:04:05")
		m := fmt.Sprintf("[ws][%s]:%s",tm,msg)
		err = ws.WriteMessage(1,[]byte(m))

		if err != nil{
			fmt.Println(err)
		}
	}
}

// func sendUserMsg(ctx *gin.Context){
// 	ws,err := upGrade.Upgrade(ctx.Writer,ctx.Request,nil)
// 	if err != nil{
// 		fmt.Println(err)
// 		return
// 	}

// 	defer func(ws *websocket.Conn){
// 		err = ws.Close()
// 		if err != nil{
// 			fmt.Println(err)
// 		}
// 	}(ws)

// 	MsgHandler(ws,ctx)
// }

func SendUserMsg(ctx *gin.Context){
	models.Chat(ctx.Writer,ctx.Request)
}