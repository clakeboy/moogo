package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/clakeboy/golib/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"moogo/common"
	"sort"
)

//列表显示字段名
type MongoColumn struct {
	Key      string        `json:"key"`
	Type     bsontype.Type `json:"type"`
	TypeName string        `json:"type_name"`
}

type ColumnList []*MongoColumn

func (l ColumnList) Len() int           { return len(l) }
func (l ColumnList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l ColumnList) Less(i, j int) bool { return l[i].Key < l[j].Key }

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

	coll := conn.Db.Database(params.Database).Collection(params.Collection)

	var dataCount int64
	if len(filter) > 0 {
		dataCount, err = coll.CountDocuments(common.GetContent(), filter)
	} else {
		dataCount, err = coll.EstimatedDocumentCount(common.GetContent())
	}
	if err != nil {
		return nil, err
	}

	findOpt := options.Find()
	findOpt.SetLimit(params.Number)
	findOpt.SetSkip((params.Page - 1) * params.Number)
	findOpt.SetSort(mSort)
	//findOpt.SetProjection()
	cur, err := coll.Find(common.GetContent(),
		filter,
		findOpt,
	)

	if err != nil {
		return nil, err
	}

	defer cur.Close(common.GetContent())

	var list []interface{}
	keysM := utils.M{}
	var keys ColumnList
	for cur.Next(common.GetContent()) {
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
			keys = append(keys, &MongoColumn{
				Key:      v.Key(),
				Type:     v.Value().Type,
				TypeName: v.Value().Type.String(),
			})
		}

		list = append(list, val)
	}

	sort.Sort(keys)

	return utils.M{
		"list":  list,
		"count": dataCount,
		"keys":  keys,
	}, nil
}

//得到一条记录
func (e *ExecController) ActionFind(args []byte) (string, error) {
	var params struct {
		ServerId   int    `json:"server_id"`
		Database   string `json:"database"`
		Collection string `json:"collection"`
		Id         string `json:"id"`
		IdType     int    `json:"id_type"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return "", err
	}
	var id interface{}
	if params.IdType == 7 {
		id, err = primitive.ObjectIDFromHex(params.Id)
		if err != nil {
			return "", err
		}
	} else {
		id = params.Id
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return "", err
	}

	coll := conn.Db.Database(params.Database).Collection(params.Collection)

	res := coll.FindOne(common.GetContent(), bson.M{"_id": id})
	if res.Err() != nil {
		return "", res.Err()
	}

	row, err := res.DecodeBytes()
	if err != nil {
		return "", err
	}

	data, err := bson.MarshalExtJSON(row, true, false)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

//添加数据
func (e *ExecController) ActionInsert(args []byte) error {
	var params struct {
		ServerId   int    `json:"server_id"`
		Database   string `json:"database"`
		Collection string `json:"collection"`
		Data       string `json:"data"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	data := bson.M{}
	err = bson.UnmarshalExtJSON([]byte(params.Data), true, &data)
	if err != nil {
		return err
	}

	data["_id"] = primitive.NewObjectID()

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return err
	}

	_, err = conn.Db.Database(params.Database).Collection(params.Collection).InsertOne(common.GetContent(), data)

	return err
}

//删除
func (e *ExecController) ActionDelete(args []byte) error {
	var params struct {
		ServerId   int    `json:"server_id"`
		Database   string `json:"database"`
		Collection string `json:"collection"`
		Id         string `json:"id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	id, err := primitive.ObjectIDFromHex(params.Id)
	if err != nil {
		return err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return err
	}

	coll := conn.Db.Database(params.Database).Collection(params.Collection)
	_, err = coll.DeleteOne(common.GetContent(), bson.M{"_id": id})

	return err
}

//删除多项
func (e *ExecController) ActionDeleteMulti(args []byte) (int, error) {
	var params struct {
		ServerId   int      `json:"server_id"`
		Database   string   `json:"database"`
		Collection string   `json:"collection"`
		IdList     []string `json:"id_list"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return 0, err
	}

	var ids []interface{}
	for _, v := range params.IdList {
		id, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return 0, fmt.Errorf("%s,err:%s", id, err.Error())
		}
		ids = append(ids, id)
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return 0, err
	}

	coll := conn.Db.Database(params.Database).Collection(params.Collection)
	res, err := coll.DeleteMany(common.GetContent(), bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return 0, err
	}
	return int(res.DeletedCount), err
}

//修改
func (e *ExecController) ActionUpdate(args []byte) error {
	var params struct {
		ServerId   int    `json:"server_id"`
		Database   string `json:"database"`
		Collection string `json:"collection"`
		Data       string `json:"data"`
		Id         string `json:"id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	id, err := primitive.ObjectIDFromHex(params.Id)
	if err != nil {
		return err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return err
	}
	data := bson.M{}
	err = bson.UnmarshalExtJSON([]byte(params.Data), true, &data)
	if err != nil {
		return err
	}

	coll := conn.Db.Database(params.Database).Collection(params.Collection)
	_, err = coll.ReplaceOne(common.GetContent(), bson.M{"_id": id}, data)
	return err
}
