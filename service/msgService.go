package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hxysj/im-system/models"
	"github.com/hxysj/im-system/utils"
)

// 防止跨域站点伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) {
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ws)

	MsgHandler(ws, c)
}

func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	for {
		msg, err := utils.Subscribe(c, utils.PublishKey)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("发送消息 ", msg)
		tm := time.Now().Format("2006-01-02 15:04:05")
		m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
		err = ws.WriteMessage(1, []byte(m))

		if err != nil {
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

func Chat(ctx *gin.Context) {
	userId := ctx.GetInt64("current_user_id")
	if userId <= 0 {
		return
	}
	models.Chat(userId, ctx.Writer, ctx.Request)
}

type EmojiInfo struct {
	Emojis []Emoji `json:"emojis"`
}

type Emoji struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	File string `json:"file"`
}

// GetEmojiList
// Summary 获取表情包
// @Tags 消息模块
// Success 200 {string} json{"code","message"}
// @Router /msg/getEmojiList [get]
func GetEmojiList(ctx *gin.Context) {
	// 读取文件，获取文件内容
	content, err := os.ReadFile("asset/emoji/info.json")
	if err != nil {
		fmt.Println(err)
		utils.RespFail(ctx.Writer, "获取表情包失败！")
		return
	}
	var Result EmojiInfo
	err = json.Unmarshal(content, &Result)
	if err != nil {
		fmt.Println(err)
		utils.RespFail(ctx.Writer, "获取表情包失败！")
		return
	}

	for index := range Result.Emojis {
		Result.Emojis[index].File = "/asset/emoji/" + Result.Emojis[index].File
	}
	utils.RespOkList(ctx.Writer, Result.Emojis, len(Result.Emojis))
}

// GetMessageList
// Summary 获取消息
// @Tags 消息模块
// @Param conversation_id query string false `会话的id`
// @Param limit query string false `消息数量`
// @Param page query string false `页码`
func GetMessageList(ctx *gin.Context) {
	userId := ctx.GetInt64("current_user_id")
	conversationId, conErr := strconv.Atoi(ctx.Query("conversation_id"))
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "30"))

	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))

	if err != nil || conErr != nil {
		utils.RespFail(ctx.Writer, "参数有误")
		return
	}

	if conversationId <= 0 || limit <= 0 || limit > 100 || page <= 0 {
		utils.RespFail(ctx.Writer, "参数有误")
		return
	}

	list, total, err := models.GetMessageList(userId, int64(conversationId), limit, page)

	if err != nil {
		fmt.Println(err)
		utils.RespFail(ctx.Writer, "获取消息列表失败")
	} else {
		utils.RespOkList(ctx.Writer, list, total)
	}
}

func ReadMessage(ctx *gin.Context) {
	user_id := ctx.GetInt64("current_user_id")
	conversation_id, err := strconv.Atoi(ctx.PostForm("conversation_id"))

	if err != nil || conversation_id <= 0 {
		utils.RespFail(ctx.Writer, "参数有误！")
		return
	}

	resErr := models.ReadMessage(user_id, int64(conversation_id))
	if resErr != nil {
		utils.RespFail(ctx.Writer, "设置失败！")
	} else {
		utils.RespOk(ctx.Writer, nil, "设置成功！")
	}
}
