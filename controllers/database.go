package controllers

import (
	"encoding/json"
	"github.com/clakeboy/golib/utils"
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
func (d *DatabaseController) ActionQuery(args []byte) ([]utils.M, error) {
	var params struct {
		ServerId int `json:"server_id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return nil, err
	}

	result, err := conn.Db.ListDatabase()
	if err != nil {
		return nil, err
	}

	var list []utils.M

	for _, v := range result.Databases {
		list = append(list, utils.M{
			"name": v.Name,
		})
	}

	return list, nil
}

//添加
func (d *DatabaseController) ActionAdd(args []byte) (string, error) {
	var params struct {
		ServerId int    `json:"server_id"`
		Database string `json:"database"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return "", err
	}

	//conn, err := common.Conns.Get(params.ServerId)
	//if err != nil {
	//	return "", err
	//}
	//
	//coll := conn.Db.Database(params.Database).Collection("tmp")
	//_, err = coll.InsertOne(common.GetContent(), bson.M{"name": params.Database})
	//if err != nil {
	//	return "", nil
	//}
	//_ = coll.Drop(nil)
	return params.Database, nil
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

	err = conn.Db.Database(params.Database).Drop(common.GetContent())

	return err
}
