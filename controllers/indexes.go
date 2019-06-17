package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/clakeboy/golib/ckdb"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"moogo/common"
)

//索引管理
type IndexesController struct {
	c *gin.Context
}

func NewIndexesController(c *gin.Context) *IndexesController {
	return &IndexesController{c: c}
}

//查询
func (i *IndexesController) ActionQuery(args []byte) (*ckdb.QueryResult, error) {
	var params struct {
		Page   int `json:"page"`
		Number int `json:"number"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

//查找一个索引
func (i *IndexesController) ActionFind(args []byte) (bson.M, error) {
	var params struct {
		ServerId   int    `json:"server_id"`
		Database   string `json:"database"`
		Collection string `json:"collection"`
		IndexName  string `json:"index_name"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return nil, err
	}
	ctx := context.TODO()
	idxView := conn.Db.Database(params.Database).Collection(params.Collection).Indexes()
	cur, err := idxView.List(ctx)
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var data bson.M
		err := cur.Decode(&data)
		if err != nil {
			continue
		}
	}

	return nil, nil
}

//添加
func (i *IndexesController) ActionAdd(args []byte) error {
	var params struct {
		ServerId   int             `json:"server_id"`
		Database   string          `json:"database"`
		Collection string          `json:"collection"`
		Keys       json.RawMessage `json:"keys"`
		Opts       struct {
			Name               string `json:"name"`                 //索引名称
			Background         bool   `json:"background"`           //是否后端执行索引
			Unique             bool   `json:"unique"`               //是否唯一键
			ExpireAfterSeconds int32  `json:"expire_after_seconds"` //索引过期时间
		} `json:"opts"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return err
	}

	keys := bson.M{}

	err = bson.UnmarshalExtJSON(params.Keys, true, &keys)

	if err != nil {
		return err
	}

	idxOpts := options.Index()
	idxOpts.SetName(params.Opts.Name)
	idxOpts.SetBackground(params.Opts.Background)
	idxOpts.SetExpireAfterSeconds(params.Opts.ExpireAfterSeconds)
	idxOpts.SetUnique(params.Opts.Unique)

	ctx := context.TODO()
	dbIdx := conn.Db.Database(params.Database).Collection(params.Collection).Indexes()
	res, err := dbIdx.CreateOne(ctx, mongo.IndexModel{
		Keys:    keys,
		Options: idxOpts,
	})
	fmt.Println(res)
	return err
}

//删除
func (i *IndexesController) ActionDelete(args []byte) error {
	return nil
}

//修改
func (i *IndexesController) ActionUpdate(args []byte) error {
	return nil
}
