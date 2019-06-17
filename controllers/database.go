package controllers

import (
	"context"
	"encoding/json"
	"github.com/clakeboy/golib/ckdb"
	"github.com/gin-gonic/gin"
	"moogo/common"
)

//数据库管理
type DatabaseController struct {
	c *gin.Context
}

func NewDatabaseController(c *gin.Context) *DatabaseController {
	return &DatabaseController{c: c}
}

//查询
func (d *DatabaseController) ActionQuery(args []byte) (*ckdb.QueryResult, error) {
	var params struct {
		ServerId int `json:"server_id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

//添加
func (d *DatabaseController) ActionAdd(args []byte) error {
	var params struct {
		ServerId int    `json:"server_id"`
		Database string `json:"database"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return err
	}

	_ = conn.Db.Database(params.Database)

	return nil
}

//删除
func (d *DatabaseController) ActionDrop(args []byte) error {
	var params struct {
		ServerId int    `json:"server_id"`
		Database string `json:"database"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return err
	}

	err = conn.Db.Database(params.Database).Drop(context.TODO())

	return err
}
