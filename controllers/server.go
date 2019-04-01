package controllers

import (
	"encoding/json"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/clakeboy/golib/ckdb"
	"github.com/gin-gonic/gin"
	"moogo/models"
)

//服务器控制器
type ServerController struct {
	c *gin.Context
}

//创建服务器控制器
func NewServerController(c *gin.Context) *ServerController {
	return &ServerController{
		c: c,
	}
}

//查询服务器列表
func (s *ServerController) ActionQuery(args []byte) (*ckdb.QueryResult, error) {
	var params struct {
		Name   string `json:"name"`
		Page   int    `json:"page"`
		Number int    `json:"number"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	var where []q.Matcher
	if params.Name != "" {
		where = append(where, q.Re("Name", params.Name))
	}

	model := models.NewServerConnect(nil)
	res, err := model.Query(params.Page, params.Number, where...)
	if err != nil && err != storm.ErrNotFound {
		return nil, nil
	}

	return res, nil
}

//修改和添加记录
func (s *ServerController) ActionEdit(args []byte) error {
	var data models.ServerConnectData
	err := json.Unmarshal(args, &data)
	if err != nil {
		return err
	}

	model := models.NewServerConnect(nil)
	if data.Id == 0 {
		return model.Save(&data)
	}
	return model.Update(&data)
}
