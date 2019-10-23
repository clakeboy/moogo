package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/clakeboy/golib/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"moogo/common"
	"sort"
)

//控制器
type CollectionController struct {
	c *gin.Context
}

func NewCollectionController(c *gin.Context) *CollectionController {
	return &CollectionController{c: c}
}

//得到所有文档集合列表
func (c *CollectionController) ActionList(args []byte) ([]utils.M, error) {
	var params struct {
		ServerId int    `json:"server_id"`
		Database string `json:"database"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return nil, err
	}

	list, err := conn.Db.Database(params.Database).ListCollectionNames(nil, bson.D{})
	sort.Strings(list)
	var res []utils.M
	for _, v := range list {
		res = append(res, utils.M{
			"collection": v,
		})
	}
	return res, err
}

//删除一个文档集合
func (c *CollectionController) ActionDelete(args []byte) error {
	var params struct {
		ServerId   int    `json:"server_id"`
		Database   string `json:"database"`
		Collection string `json:"collection"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return err
	}

	coll := conn.Db.Database(params.Database).Collection(params.Collection)
	err = coll.Drop(context.TODO())
	return err
}

//创建一个文档集合
func (c *CollectionController) ActionCreate(args []byte) error {
	var params struct {
		ServerId   int    `json:"server_id"`
		Database   string `json:"database"`
		Collection string `json:"collection"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return err
	}

	db := conn.Db.Database(params.Database)
	cur := db.RunCommand(nil, bson.M{
		"create": params.Collection,
	})
	res := bson.M{}
	err = cur.Decode(&res)
	if err != nil {
		return err
	}
	if _, ok := res["ok"]; !ok {
		return errors.New("create collection error")
	}
	return nil
}
