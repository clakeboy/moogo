package models

import (
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/clakeboy/golib/ckdb"
	"moogo/common"
)

//服务器连接模型
type ServerConnect struct {
	Table string `json:"table"` //表名
	storm.Node
}

func NewServerConnect(db *storm.DB) *ServerConnect {
	if db == nil {
		db = common.BDB
	}

	return &ServerConnect{
		Table: "server",
		Node:  db.From("server"),
	}
}

//通过ID拿到记录
func (s *ServerConnect) GetById(id int) (*common.ServerConnectData, error) {
	data := &common.ServerConnectData{}
	err := s.One("Id", id, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

//查询条件得到任务数据列表
func (s *ServerConnect) Query(page, number int, where ...q.Matcher) (*ckdb.QueryResult, error) {
	var list []common.ServerConnectData

	count, err := s.Select(where...).Count(new(common.ServerConnectData))
	if err != nil {
		return nil, err
	}
	//fmt.Println(s.Select(where...).Find(&list))
	//fmt.Println(list)

	err = s.Select(where...).Find(&list)
	if err != nil {
		return nil, err
	}
	return &ckdb.QueryResult{
		List:  list,
		Count: count,
	}, nil
}

//查询条件得到任务数据列表
func (s *ServerConnect) List(page, number int, where ...q.Matcher) ([]common.ServerConnectData, error) {
	var list []common.ServerConnectData
	err := s.Select(where...).Limit(number).Skip((page - 1) * number).Find(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//指定条件删除记录
func (s *ServerConnect) Remove(where ...q.Matcher) error {
	if len(where) <= 0 {
		return errors.New("must be one condition")
	}

	return s.Select(where...).Delete(new(common.ServerConnectData))
}
