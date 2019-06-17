package models

import (
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/clakeboy/golib/ckdb"
	"moogo/common"
)

type FormatData struct {
	Id          int    `storm:"id,increment" json:"id"` //主键,自增长
	Name        string `json:"name" storm:"unique"`     //format方法名
	Content     string `json:"content"`                 //format内容
	CreatedDate int    `json:"created_date"`            //创建时间
}

//服务器连接模型
type FormatModal struct {
	Table string `json:"table"` //表名
	storm.Node
}

func NewFormatModal(db *storm.DB) *FormatModal {
	if db == nil {
		db = common.BDB
	}

	return &FormatModal{
		Table: "format",
		Node:  db.From("format"),
	}
}

//通过ID拿到记录
func (s *FormatModal) GetById(id int) (*FormatData, error) {
	data := &FormatData{}
	err := s.One("Id", id, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

//查询条件得到任务数据列表
func (s *FormatModal) Query(page, number int, where ...q.Matcher) (*ckdb.QueryResult, error) {
	var list []FormatData

	count, err := s.Select(where...).Count(new(FormatData))
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
func (s *FormatModal) List(page, number int, where ...q.Matcher) ([]FormatData, error) {
	var list []FormatData
	err := s.Select(where...).Limit(number).Skip((page - 1) * number).Find(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//指定条件删除记录
func (s *FormatModal) Remove(where ...q.Matcher) error {
	if len(where) <= 0 {
		return errors.New("must be one condition")
	}

	return s.Select(where...).Delete(new(FormatData))
}
