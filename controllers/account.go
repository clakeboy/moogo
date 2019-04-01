package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/asdine/storm/q"
	"github.com/clakeboy/golib/ckdb"
	"github.com/clakeboy/golib/utils"
	"github.com/gin-gonic/gin"
	"moogo/models"
)

//控制器
type AccountController struct {
	c *gin.Context
}

func NewAccountController(c *gin.Context) *AccountController {
	return &AccountController{c: c}
}

//查询
func (a *AccountController) ActionQuery(args []byte) (*ckdb.QueryResult, error) {
	var params struct {
		Account string `json:"account"`
		Page    int    `json:"page"`
		Number  int    `json:"number"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	var where []q.Matcher
	if params.Account != "" {
		where = append(where, q.Re("Account", params.Account))
	}

	model := models.NewAccountModel(nil)
	res, err := model.Query(params.Page, params.Number, where...)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//添加
func (a *AccountController) ActionInsert(args []byte) error {
	var params struct {
		Account  string `json:"account"`   //帐户名
		UserName string `json:"user_name"` //用户姓名
		Password string `json:"password"`  //用户密码
		Disable  bool   `json:"disable"`   //用户是否禁用
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	model := models.NewAccountModel(nil)

	data := &models.AccountData{
		Account:     params.Account,
		UserName:    params.UserName,
		Password:    utils.EncodeMD5(params.Password),
		Disable:     params.Disable,
		CreatedDate: int(time.Now().Unix()),
	}

	err = model.Save(data)
	if err != nil {
		return utils.Error("添加用户出错!", err)
	}

	return nil
}

//使用ID得到一条记录
func (m *AccountController) ActionFind(args []byte) (*models.AccountData, error) {
	var params struct {
		Id int `json:"id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	model := models.NewAccountModel(nil)
	data, err := model.GetById(params.Id)
	if err != nil {
		return nil, err
	}

	data.Password = ""

	return data, nil
}

//删除
func (a *AccountController) ActionDisable(args []byte) error {
	var params struct {
		Id int `json:"id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	return nil
}

//修改
func (a *AccountController) ActionUpdate(args []byte) error {
	var params struct {
		Id       int    `json:"id"`
		Account  string `json:"account"`   //帐户名
		UserName string `json:"user_name"` //用户姓名
		Password string `json:"password"`  //用户密码
		Disable  bool   `json:"disable"`   //用户是否禁用
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	model := models.NewAccountModel(nil)

	data, err := model.GetById(params.Id)
	if err != nil {
		return utils.Error("查询用户出错!", err)
	}

	data.UserName = params.UserName
	if params.Password != "" {
		data.Password = utils.EncodeMD5(params.Password)
	}
	data.Disable = params.Disable

	fmt.Println(data)

	err = model.Update(data)
	if err != nil {
		return utils.Error("修改用户出错!", err)
	}

	return nil
}
