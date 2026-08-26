package service

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/utils"
)

func Upload(ctx *gin.Context) {
	w := ctx.Writer
	req := ctx.Request
	// 获取用户上传的文件数据
	file, header, err := req.FormFile("file")

	if err != nil {
		fmt.Println(err)
		utils.RespFail(w, "上传文件失败！")
		return
	}
	defer file.Close()
	suffix := "png"
	// 获取上传的文件名称
	originFileName := header.Filename
	tem := strings.Split(originFileName, ".")
	// 获取文件后缀
	if len(tem) > 1 {
		suffix = "." + tem[len(tem)-1]
	}
	// 重新生成唯一的文件名称
	fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), suffix)
	// 创建空文件
	dstFile, err := os.Create("./asset/upload/" + fileName)

	if err != nil {
		fmt.Println(err)
		utils.RespFail(w, "上传文件失败！")
		return
	}
	// 给创建的文件赋值
	defer dstFile.Close()
	_, err = io.Copy(dstFile, file)
	if err != nil {
		fmt.Println(err)
		utils.RespFail(w, "上传文件失败！")
		return
	}
	// 返回文件链接
	url := "/asset/upload/" + fileName
	utils.RespOk(w, url, "上传文件成功！")
}
