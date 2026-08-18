package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

// 
func Chat(ctx *gin.Context){
	models.Chat(ctx.Writer,ctx.Request)
}

type EmojiInfo struct{
	Emojis []Emoji `json:"emojis"`
}

type Emoji struct{
	ID int `json:"id"`
	Name string `json:"name"`
	File string `json:"file"`
}

// GetEmojiList
// Summary 获取表情包
// @Tags 消息模块
// Success 200 {string} json{"code","message"}
// @Router /msg/getEmojiList [get]
func GetEmojiList(ctx *gin.Context){
	// 读取文件，获取文件内容
	content,err := os.ReadFile("asset/emoji/info.json")
	if err !=nil{
		fmt.Println(err)
		utils.RespFail(ctx.Writer,"获取表情包失败！")
		return
	}
	var Result EmojiInfo
	err = json.Unmarshal(content,&Result)
	if err !=nil{
		fmt.Println(err)
		utils.RespFail(ctx.Writer,"获取表情包失败！")
		return
	}

	for index := range Result.Emojis{
		Result.Emojis[index].File = "/asset/emoji/" + Result.Emojis[index].File
	}
	utils.RespOkList(ctx.Writer,Result.Emojis,len(Result.Emojis))
}

// GetMessageList
// Summary 获取用户私聊消息
// @Tags 消息模块
// @Param user_id query string false `用户id`
// @Param target_id query string false `好友id`
// @Param limit query string false `消息数量`
// @Param page query string false `页码`
// func GetMessageList(ctx *gin.Context){
// 	userId,_ := strconv.Atoi(ctx.Query("user_id"))
// 	targetId,_ := strconv.Atoi(ctx.Query("target_id"))
// 	limit,_ := strconv.Atoi(ctx.Query("limit"))
// 	page,_ := strconv.Atoi(ctx.Query("page"))
// }