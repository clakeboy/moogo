package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
)

//查询结果
type QueryResult struct {
	List  interface{} `json:"list"`
	Count int         `json:"count"`
}

type Collection struct {
	collectionName string
	Coll           *mongo.Collection
}

func NewCollection(collectionName string, coll *mongo.Collection) *Collection {
	return &Collection{
		collectionName: collectionName,
		Coll:           coll,
	}
}

func (c *Collection) Insert(rows ...interface{}) error {
	ctx := context.Background()
	_, err := c.Coll.InsertMany(ctx, rows)
	return err
}

func (c *Collection) Update(where bson.M, update bson.M) error {
	ctx := context.Background()
	_, err := c.Coll.UpdateMany(ctx, where, update)
	return err
}

func (c *Collection) Query(where bson.M, page int64, number int64, sort bson.M, structType interface{}) (*QueryResult, error) {
	ctx := context.Background()
	dataCount, err := c.Count(where)
	if err != nil {
		return nil, err
	}

	findOpt := options.Find()
	if number >= 1 {
		findOpt.SetLimit(number)
	}
	if page >= 1 {
		findOpt.SetSkip((page - 1) * number)
	}
	if sort != nil {
		findOpt.SetSort(sort)
	}

	cur, err := c.Coll.Find(ctx,
		where,
		findOpt,
	)
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)
	var list []interface{}
	for cur.Next(ctx) {
		result := c.getQueryType(structType)
		err = cur.Decode(result)
		if err != nil {
			return nil, err
		}
		list = append(list, result)
	}

	return &QueryResult{list, int(dataCount)}, nil
}

func (c *Collection) List(where bson.M, page int64, number int64, sort bson.M, structType interface{}) ([]interface{}, error) {
	ctx := context.Background()

	findOpt := options.Find()
	if number >= 1 {
		findOpt.SetLimit(number)
	}
	if page >= 1 {
		findOpt.SetSkip((page - 1) * number)
	}
	if sort != nil {
		findOpt.SetSort(sort)
	}

	cur, err := c.Coll.Find(ctx,
		where,
		findOpt,
	)
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)
	var list []interface{}
	for cur.Next(ctx) {
		result := c.getQueryType(structType)
		err = cur.Decode(result)
		if err != nil {
			return nil, err
		}
		list = append(list, result)
	}

	return list, nil
}

func (c *Collection) getQueryType(i interface{}) interface{} {
	if i == nil {
		return &bson.M{}
	}

	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface()
	}
	return reflect.New(t).Interface()
}

//
func (c *Collection) Count(where bson.M) (int64, error) {
	ctx := context.Background()
	if where != nil && len(where) > 0 {
		return c.Coll.CountDocuments(ctx, where)
	} else {
		return c.Coll.EstimatedDocumentCount(ctx)
	}
}

//克隆一个连接
func (c *Collection) Clone() (*Collection, error) {
	coll, err := c.Coll.Clone()
	if err != nil {
		return nil, err
	}
	return NewCollection(c.collectionName, coll), nil
}
