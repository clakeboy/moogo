package controllers

import (
	"encoding/json"
	"strconv"

	"github.com/clakeboy/golib/httputils"
	"github.com/clakeboy/golib/utils"
	"github.com/gin-gonic/gin"
	"moogo/models"
)

//控制器
type LoginController struct {
	c          *gin.Context
	cookieName string
}

func NewLoginController(c *gin.Context) *LoginController {
	return &LoginController{
		c:          c,
		cookieName: "moogo",
	}
}

//查询
func (l *LoginController) ActionSign(args []byte) (*models.AccountData, error) {
	var params struct {
		Account  string `json:"account"`
		Password string `json:"password"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	model := models.NewAccountModel(nil)
	data, err := model.GetByAccount(params.Account)
	if err != nil {
		return nil, utils.Error("没有找到用户!", err)
	}

	if data.Password != utils.EncodeMD5(params.Password) {
		return nil, utils.Error("用户名或密码错误!", nil)
	}

	if data.Disable {
		return nil, utils.Error("用户已经被禁用!", nil)
	}

	cookie := l.c.MustGet("cookie").(*httputils.HttpCookie)

	cookie.Set(l.cookieName, strconv.Itoa(data.Id), 3600*24*365)

	data.Password = ""

	return data, nil
}

//是否登录验证
func (l *LoginController) ActionAuth(args []byte) (*models.AccountData, error) {
	cookie := l.c.MustGet("cookie").(*httputils.HttpCookie)
	id, err := cookie.Get(l.cookieName)
	if err != nil {
		return nil, utils.Error("auth failed", nil)
	}

	aid, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	model := models.NewAccountModel(nil)
	data, err := model.GetById(aid)
	if err != nil {
		return nil, err
	}

	data.Password = ""

	return data, nil
}

//退出登录
func (l *LoginController) ActionLogout(args []byte) error {
	cookie := l.c.MustGet("cookie").(*httputils.HttpCookie)
	cookie.Delete(l.cookieName)
	return nil
}
