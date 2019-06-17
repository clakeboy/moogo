package controllers

import (
	"context"
	"encoding/json"
	"github.com/clakeboy/golib/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"moogo/common"
	"time"
)

//控制器
type ExecController struct {
	c *gin.Context
}

func NewExecController(c *gin.Context) *ExecController {
	return &ExecController{c: c}
}

//查询
func (e *ExecController) ActionQuery(args []byte) (utils.M, error) {
	var params struct {
		ServerId   int             `json:"server_id"`
		Database   string          `json:"database"`
		Collection string          `json:"collection"`
		Filter     json.RawMessage `json:"filter"`
		Sort       json.RawMessage `json:"sort"`
		Page       int64           `json:"page"`
		Number     int64           `json:"number"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return nil, err
	}
	filter := bson.D{}
	err = bson.UnmarshalExtJSON(params.Filter, true, &filter)
	if err != nil {
		return nil, err
	}

	mSort := bson.M{}
	err = bson.UnmarshalExtJSON(params.Sort, true, &mSort)
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()

	coll := conn.Db.Database(params.Database).Collection(params.Collection)

	var dataCount int64
	if len(filter) > 0 {
		dataCount, err = coll.CountDocuments(ctx, filter)
	} else {
		dataCount, err = coll.EstimatedDocumentCount(ctx)
	}
	if err != nil {
		return nil, err
	}

	findOpt := options.Find()
	findOpt.SetLimit(params.Number)
	findOpt.SetSkip((params.Page - 1) * params.Number)
	findOpt.SetSort(mSort)
	cur, err := coll.Find(ctx,
		filter,
		findOpt,
	)

	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	var list []interface{}
	keysM := utils.M{}
	var keys []utils.M
	for cur.Next(ctx) {
		//data := bson.M{}
		val := bson.M{}
		data := bson.Raw{}
		_ = cur.Decode(&data)
		_ = cur.Decode(&val)
		elms, _ := data.Elements()
		for _, v := range elms {
			if _, ok := keysM[v.Key()]; ok {
				continue
			}
			keysM[v.Key()] = true
			keys = append(keys, utils.M{
				"key":       v.Key(),
				"type":      v.Value().Type,
				"type_name": v.Value().Type.String(),
			})
		}
		list = append(list, val)
	}

	return utils.M{
		"list":  list,
		"count": dataCount,
		"keys":  keys,
	}, nil
}

//添加数据
func (e *ExecController) ActionInsert(args []byte) error {
	var params struct {
		ServerId   int             `json:"server_id"`
		Database   string          `json:"database"`
		Collection string          `json:"collection"`
		Data       json.RawMessage `json:"data"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	data := bson.M{}
	err = bson.UnmarshalExtJSON(params.Data, true, &data)
	if err != nil {
		return err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return err
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)

	_, err = conn.Db.Database(params.Database).Collection(params.Collection).InsertOne(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

//删除
func (e *ExecController) ActionDelete(args []byte) error {
	var params struct {
		ServerId   int             `json:"server_id"`
		Database   string          `json:"database"`
		Collection string          `json:"collection"`
		Data       json.RawMessage `json:"data"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}
	return nil
}

//修改
func (e *ExecController) ActionUpdate(args []byte) error {
	var params struct {
		ServerId   int             `json:"server_id"`
		Database   string          `json:"database"`
		Collection string          `json:"collection"`
		Data       json.RawMessage `json:"data"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}
	return nil
}
