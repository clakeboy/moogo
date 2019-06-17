package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/gin-gonic/gin"
	"moogo/common"
	"moogo/components/mongo"
	"moogo/models"
	"time"
)

//服务器控制器
type ServerController struct {
	c *gin.Context
}

//创建服务器控制器
func NewServerController(c *gin.Context) *ServerController {
	return &ServerController{
		c: c,
	}
}

//查询服务器列表
func (s *ServerController) ActionQuery(args []byte) ([]common.ServerConnectData, error) {
	var params struct {
		Name   string `json:"name"`
		Page   int    `json:"page"`
		Number int    `json:"number"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	var where []q.Matcher
	if params.Name != "" {
		where = append(where, q.Re("Name", params.Name))
	}

	model := models.NewServerConnect(nil)
	res, err := model.List(params.Page, params.Number, where...)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	return res, nil
}

//修改和添加记录
func (s *ServerController) ActionEdit(args []byte) error {
	var data common.ServerConnectData
	err := json.Unmarshal(args, &data)
	if err != nil {
		return err
	}

	model := models.NewServerConnect(nil)
	if data.Id == 0 {
		data.CreatedDate = int(time.Now().Unix())
	}
	return model.Save(&data)
	//return model.Update(&data)
}

//得到一条记录
func (s *ServerController) ActionFind(args []byte) (*common.ServerConnectData, error) {
	var params struct {
		ServerID int `json:"id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	model := models.NewServerConnect(nil)
	data, err := model.GetById(params.ServerID)
	return data, err
}

//删除一条记录
func (s *ServerController) ActionDelete(args []byte) error {
	var params struct {
		ServerID int `json:"id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	model := models.NewServerConnect(nil)
	return model.DeleteStruct(&common.ServerConnectData{
		Id: params.ServerID,
	})
}

//测试连接
func (s *ServerController) ActionTestConnect(args []byte) error {
	var params struct {
		ServerID int `json:"id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	model := models.NewServerConnect(nil)
	serverInfo, err := model.GetById(params.ServerID)
	if err != nil {
		return err
	}
	cfg := &mongo.Config{
		Host:     serverInfo.Address,
		Port:     serverInfo.Port,
		PoolSize: 1,
	}
	if serverInfo.IsAuth {
		cfg.Auth = serverInfo.AuthDatabase
		cfg.User = serverInfo.AuthUser
		cfg.Password = serverInfo.AuthPassword
	}

	var sshSess *common.SSHSession
	if serverInfo.IsSSH {
		client, err := common.LoginSSH(&common.SSHServer{
			Addr:     fmt.Sprintf("%s:%s", serverInfo.SSHAddress, serverInfo.SSHPort),
			User:     serverInfo.SSHUser,
			Password: serverInfo.SSHPassword,
		})
		if err != nil {
			fmt.Println("login ssh server error: ", err)
			return err
		}
		sshSess = common.NewSession(":33001",
			fmt.Sprintf("%s:%s", serverInfo.Address, serverInfo.Port),
			client)
		go sshSess.Run()
		cfg.Host = "127.0.0.1"
		cfg.Port = "33001"
	}

	db, err := mongo.NewDatabase(cfg)
	if err != nil {
		return err
	}

	err = db.Open()
	if err != nil {
		return err
	}
	if serverInfo.IsSSH {
		sshSess.Close()
	}

	return nil
}
