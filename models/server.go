package models

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/clakeboy/golib/ckdb"
	"moogo/common"
)

//服务器连接数据
type ServerConnectData struct {
	Id            int    `storm:"id,increment" json:"id"` //主键,自增长
	Name          string `json:"name" storm:"unique"`     //服务器名称
	Address       string `json:"address"`                 //服务器IP地址
	Port          string `json:"port"`                    //服务器端口号
	IsAuth        int    `json:"is_auth"`                 //是否验证用户名,0,1
	AuthDatabase  string `json:"auth_database"`           //验证用户数据库名
	AuthUser      string `json:"auth_user"`               //验证用户名
	AuthPassword  string `json:"auth_password"`           //验证用户密码
	IsSSH         int    `json:"is_ssh"`                  //是否使用SSH,0,1
	SSHAddress    string `json:"ssh_address"`             //SSH服务IP地址
	SSHPort       string `json:"ssh_port"`                //SSH服务端口号
	SSHUser       string `json:"ssh_user"`                //SSH服务用户名
	SSHAuthMethod string `json:"ssh_auth_method"`         //SSH服务验证方法, password,private key
	SSHPassword   string `json:"ssh_password"`            //SSH服务密码
	SSHKeyFile    string `json:"ssh_key_file"`            //SSH密钥文件
	CreatedDate   int    `json:"created_date"`            //创建时间
}

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
func (s *ServerConnect) GetById(id int) (*ServerConnectData, error) {
	data := &ServerConnectData{}
	err := s.One("Id", id, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

//查询条件得到任务数据列表
func (s *ServerConnect) Query(page, number int, where ...q.Matcher) (*ckdb.QueryResult, error) {
	var list []AccountData
	count, err := s.Select(where...).Count(new(AccountData))
	if err != nil {
		return nil, err
	}
	err = s.Select(where...).Limit(number).Skip((page - 1) * number).Find(&list)
	if err != nil {
		return nil, err
	}
	return &ckdb.QueryResult{
		List:  list,
		Count: count,
	}, nil
}

//查询条件得到任务数据列表
func (s *ServerConnect) List(page, number int, where ...q.Matcher) ([]ServerConnectData, error) {
	var list []ServerConnectData
	err := s.Select(where...).Limit(number).Skip((page - 1) * number).Find(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
