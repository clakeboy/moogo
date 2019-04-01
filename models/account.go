package models

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/clakeboy/golib/ckdb"
	"moogo/common"
)

//后端用户数据
type AccountData struct {
	Id          int    `storm:"id,increment" json:"id"` //主键,自增长
	Account     string `json:"account" storm:"unique"`  //帐户名
	UserName    string `json:"user_name"`               //用户姓名
	Password    string `json:"password"`                //用户密码
	Disable     bool   `json:"disable"`                 //用户是否禁用
	CreatedDate int    `json:"created_date"`            //用户创建时间
}

//后端用户模型
type AccountModel struct {
	Table string `json:"table"` //表名
	storm.Node
}

func NewAccountModel(db *storm.DB) *AccountModel {
	if db == nil {
		db = common.BDB
	}

	return &AccountModel{
		Table: "account",
		Node:  db.From("account"),
	}
}

//通过ID拿到记录
func (a *AccountModel) GetById(id int) (*AccountData, error) {
	data := &AccountData{}
	err := a.One("Id", id, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

//通过Name 拿到记录
func (t *AccountModel) GetByAccount(name string) (*AccountData, error) {
	data := &AccountData{}
	err := t.One("Account", name, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

//查询条件得到任务数据列表
func (a *AccountModel) Query(page, number int, where ...q.Matcher) (*ckdb.QueryResult, error) {
	var list []AccountData
	count, err := a.Select(where...).Count(new(AccountData))
	if err != nil {
		return nil, err
	}
	err = a.Select(where...).Limit(number).Skip((page - 1) * number).Find(&list)
	if err != nil {
		return nil, err
	}
	return &ckdb.QueryResult{
		List:  list,
		Count: count,
	}, nil
}

//查询条件得到任务数据列表
func (a *AccountModel) List(page, number int, where ...q.Matcher) ([]AccountData, error) {
	var list []AccountData
	err := a.Select(where...).Limit(number).Skip((page - 1) * number).Find(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
