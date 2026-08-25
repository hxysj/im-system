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
// @param name formData string false "用户名"
// @param password formData string false "密码"
// @param repassword formData string false "确认密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/register [Post]
func CreateUser(ctx *gin.Context){
	user := models.UserBasic{}
	user.Name = ctx.PostForm("name")
	password := ctx.PostForm("password")
	repassword := ctx.PostForm("repassword")

	if password == "" || user.Name == ""{
		ctx.JSON(200,gin.H{
			"code":-1,
			"message":"用户名或密码不能为空！",
		})
		return
	}

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
	user.UserId = utils.NextId()

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
	user.UserId = int64(id)
	code,msg := models.DeleteUser(user)

	if code == -1{
		utils.RespFail(ctx.Writer,msg)
	}else{
		utils.RespOk(ctx.Writer,nil,msg)
	}
}

// UpdateUser
// Summary 更新用户信息
// @Tags 用户模块
// @param id formData string false "用户id"
// @param name formData string false "名称"
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
	phone := ctx.PostForm("phone")
	email := ctx.PostForm("email")
	identity := ctx.PostForm("identity")

	res := models.FindUserById(id)

	if res.UserId == 0{
		ctx.JSON(200,gin.H{
			"code":-1,
			"message":"用户不存在！",
		})
		return
	}

	user.UserId = int64(id)
	user.Name = name
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
	nowTime := time.Now().Unix()
	if err := utils.DB.Model(&res).Where("user_id = ?",res.UserId).
	Updates(map[string]interface{}{
		"identity":temp,
		"login_time":uint64(nowTime),
	}).Error;err != nil{
		utils.RespFail(ctx.Writer,"登录失败！")
		return
	}

	type LoginResult struct{
		Token string `json:"token"`
		UserId int `json:"user_id"`
		Name string `json:"name"`
		Phone string `json:"phone"`
		Email string `json:"email"`
	}

	result := LoginResult{
		Token: temp,
		UserId: int(res.UserId),
		Name: res.Name,
		Phone: res.Phone,
		Email: res.Email,
	}

	utils.RespOk(ctx.Writer,result,"登录成功！")
}


// 搜索用户 - 根据手机号或者用户名称
func SearchUser(ctx *gin.Context){
	search := ctx.PostForm("keyword")

	if search == ""{
		utils.RespFail(ctx.Writer,"不能输入空的内容")
		return
	}

	res := models.FindUserByPhoneOrNameOrEmail(search)
	type UserData struct{
		UserId int64 `json:"user_id"`
		Name string `json:"name"`
		Phone string `json:"phone"`
		Email string `json:"email"`
	}
	data := UserData{
		UserId:res.UserId,
		Name:res.Name,
		Phone:res.Phone,
		Email:res.Email,
	}
	utils.RespOk(ctx.Writer,data,"")
}

// 修改密码
func UpdatePassword(ctx *gin.Context){
	user_id,_ := strconv.Atoi(ctx.PostForm("user_id"))
	oldPassword := ctx.PostForm("old_password")
	newPassword := ctx.PostForm("new_password")
	reNewPassword := ctx.PostForm("re_new_password")

	if newPassword != reNewPassword{
		utils.RespFail(ctx.Writer,"两次密码不一致")
		return
	}

	code,msg := models.UpdateUserPassword(int64(user_id),oldPassword,newPassword)

	if code == -1{
		utils.RespFail(ctx.Writer,msg)
	}else{
		utils.RespOk(ctx.Writer,nil,"修改成功！")
	}
}