package service

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/models"
	"github.com/hxysj/im-system/utils"
)

// GetUserList
// Summary 所有用户
// @Tags 用户模块
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(ctx *gin.Context){
	data := make([]models.UserBasic,10)
	data = models.GetUserList()

	ctx.JSON(200,gin.H{
		"code":0,
		"message":data,
	})

}


// CreateUser
// Summary 新增用户
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @param repassword query string false "确认密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [get]
func CreateUser(ctx *gin.Context){
	user := models.UserBasic{}
	user.Name = ctx.Query("name")
	password := ctx.Query("password")
	repassword := ctx.Query("repassword")

	salt := fmt.Sprintf("%06d",rand.Int31())

	res := models.FindUserByName(user.Name)

	if res.Name != "" {
		ctx.JSON(200,gin.H{
			"code":-1,
			"message":"用户名已被占用！",
		})
		return
	}

	res = models.FindUserByPhone(user.Phone)

	if res.Name != "" && user.Phone != ""{
		ctx.JSON(200,gin.H{
			"code":-1,
			"message":"手机号码已被占用！",
		})
		return
	}

	res = models.FindUserByEmail(user.Email)

	if res.Name != "" && user.Email != ""{
		ctx.JSON(200,gin.H{
			"code":-1,
			"message":"邮箱已被占用！",
		})
		return
	}

	if password != repassword{
		ctx.JSON(200,gin.H{
			"code":-1,
			"message":"两次密码不一致！",
		})
		return
	}

	user.Password = utils.MakePassword(password,salt)
	user.Salt =salt

	models.CreateUser(user)

	ctx.JSON(200,gin.H{
		"code":0,
		"message":"新增用户成功！",
	})

}

// DeleteUser
// Summary 删除用户
// @Tags 用户模块
// @param id query string false "用户id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [get]
func DeleteUser(ctx *gin.Context){
	user := models.UserBasic{}
	id,_ := strconv.Atoi(ctx.Query("id"))
	user.ID = uint(id)
	models.DeleteUser(user)
	ctx.JSON(200,gin.H{
		"code":0,
		"message":"删除用户成功！",
	})
}

// UpdateUser
// Summary 更新用户
// @Tags 用户模块
// @param id formData string false "用户id"
// @param name formData string false "名称"
// @param password formData string false "密码"
// @param phone formData string false "手机号"
// @param email formData string false "邮箱"
// @param identity formData string false "性别"
// @Success 200 {string} json{"code","message"}
// @Router /user/update [post]
func UpdateUser(ctx *gin.Context){
	user := models.UserBasic{}
	id,_ := strconv.Atoi(ctx.PostForm("id"))
	// 获取post请求的参数
	name := ctx.PostForm("name")
	password := ctx.PostForm("password")
	phone := ctx.PostForm("phone")
	email := ctx.PostForm("email")
	identity := ctx.PostForm("identity")

	res := models.FindUserById(id)

	if res.Name == ""{
		ctx.JSON(200,gin.H{
			"code":-1,
			"message":"用户不存在！",
		})
	}

	user.ID = uint(id)
	user.Name = name
	user.Password = utils.MakePassword(password,res.Salt)
	user.Phone = phone
	user.Email = email
	user.Identity = identity
	// 校验用户信息 - 邮箱和电话号码
	_,err := govalidator.ValidateStruct(user)
	if err != nil{
		fmt.Println(err)
		ctx.JSON(200,gin.H{
			"code":-1,
			"message":"修改参数不匹配",
		})
		return
	}

	models.UpdateUser(user)

	ctx.JSON(200,gin.H{
		"code":0,
		"message":"更新用户数据成功!",
	})
}


// Login
// Summary 登录
// @Tags 用户模块
// @param name formData string false "用户名称"
// @param password formData string false "密码"
// @Success 200 { string } json {"code","message","data"}
// @Router /user/login [post]
func Login(ctx *gin.Context){
	name := ctx.PostForm("name")
	password := ctx.PostForm("password")

	res := models.FindUserByName(name)
	if res.Name == ""{
		ctx.JSON(200,gin.H{
			"code":0,
			"message":"登录失败！",
		})
		return
	}

	validResult := utils.ValidPassword(password,res.Salt,res.Password)

	if !validResult{
		ctx.JSON(200,gin.H{
			"code":-1,
			"message":"登录失败!",
		})
		return
	}

	// 登录成功后生成token
	str := fmt.Sprintf("%d",time.Now().Unix())
	temp := utils.MD5Encode(str)
	utils.DB.Model(&res).Where("id = ?",res.ID).Update("identity",temp)

	ctx.JSON(200,gin.H{
		"message":"登录成功！",
		"data":gin.H{
			"code":0,
			"message":"登录成功！",
			"data":gin.H{
				"token":res.Identity,
				"id":res.ID,
			},
			},
	})
}
