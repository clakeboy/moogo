package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/clakeboy/golib/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"moogo/common"
	"moogo/components/mongo"
	"moogo/models"
	"sort"
)

type MenuList []*common.ServerMenu

func (l MenuList) Len() int           { return len(l) }
func (l MenuList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l MenuList) Less(i, j int) bool { return l[i].Text < l[j].Text }

//控制器
type ConnectController struct {
	c *gin.Context
}

func NewConnectController(c *gin.Context) *ConnectController {
	return &ConnectController{c: c}
}

func generateServerMenu(conn *common.Conn) ([]*common.ServerMenu, error) {
	dbList, err := conn.Db.ListDatabase()
	if err != nil {
		return nil, err
	}

	var server MenuList

	mainServer := &common.ServerMenu{
		Key:  "main",
		Text: conn.ServerInfo.Name,
		Icon: "server",
		Data: conn.ServerInfo,
		Type: "server",
	}

	var dbs MenuList

	for _, v := range dbList.Databases {
		collList, _ := generateCollectionMenu(conn, v.Name)
		dbs = append(dbs, &common.ServerMenu{
			Key:  fmt.Sprintf("db_%s", v.Name),
			Text: v.Name,
			Icon: "database",
			Data: utils.M{
				"server":        conn.ServerInfo,
				"database":      v.Name,
				"database_info": v,
			},
			Type:     "database",
			Children: collList,
		})
	}

	sort.Sort(dbs)

	mainServer.Children = dbs

	server = append(server, mainServer)

	return server, nil
}

func generateCollectionMenu(conn *common.Conn, dbName string) ([]*common.ServerMenu, error) {
	db := conn.Db.Database(dbName)
	cur, err := db.ListCollections(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	var collList MenuList

	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		coll := bson.M{}
		err := cur.Decode(&coll)
		if err != nil {
			break
		}
		collName := coll["name"].(string)
		idxList, _ := generateCollectionIndexes(conn, dbName, collName)
		collList = append(collList, &common.ServerMenu{
			Key:  fmt.Sprintf("%s_coll_%s", dbName, collName),
			Text: collName,
			Icon: "table",
			Data: utils.M{
				"server":          conn.ServerInfo,
				"database":        dbName,
				"collection":      collName,
				"collection_info": coll,
			},
			Type:     "collection",
			Children: idxList,
		})
	}

	sort.Sort(collList)

	return collList, nil
}

func generateCollectionIndexes(conn *common.Conn, dbName, collName string) ([]*common.ServerMenu, error) {
	db := conn.Db.Database(dbName)
	coll := db.Collection(collName)
	list := coll.Indexes()
	cur, err := list.List(context.TODO())
	if err != nil {
		return nil, err
	}

	var idxList MenuList

	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		idx := bson.M{}
		err := cur.Decode(&idx)
		if err != nil {
			break
		}
		idxName := idx["name"].(string)
		idxList = append(idxList, &common.ServerMenu{
			Key:  fmt.Sprintf("%s_idx_%s", collName, idxName),
			Text: idxName,
			Icon: "dice-d6",
			Data: utils.M{
				"server":     conn.ServerInfo,
				"database":   dbName,
				"collection": collName,
				"index":      idxName,
			},
			Type: "index",
		})
	}

	sort.Sort(idxList)

	return idxList, nil
}

//连接一个服务器
func (c *ConnectController) ActionConnect(args []byte) ([]*common.ServerMenu, error) {
	var params struct {
		ServerId int `json:"server_id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	if common.Conns.HasServer(params.ServerId) {
		return nil, errors.New("已经连接此服务器")
	}

	model := models.NewServerConnect(nil)
	serverInfo, err := model.GetById(params.ServerId)
	if err != nil {
		return nil, err
	}

	cfg := &mongo.Config{
		Host:     serverInfo.Address,
		Port:     serverInfo.Port,
		PoolSize: 10,
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
			return nil, err
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
		return nil, err
	}

	err = db.Open()
	if err != nil {
		return nil, err
	}

	//清除明文密码
	serverInfo.SSHPassword = ""
	serverInfo.AuthPassword = ""

	conn := &common.Conn{
		Db:         db,
		ServerInfo: serverInfo,
		SSH:        sshSess,
	}
	common.Conns.Add(conn)
	return generateServerMenu(conn)
}

//获取已经连接的服务器列表
func (c *ConnectController) ActionActiveConnect(args []byte) ([]interface{}, error) {
	if common.Conns.Len() <= 0 {
		return nil, nil
	}

	serverList := common.Conns.List()

	var menuList []interface{}

	for _, conn := range serverList {
		err := conn.Db.Ping()
		if err != nil {
			common.Conns.Remove(conn.ServerInfo.Id)
			continue
		}
		menu, _ := generateServerMenu(conn)
		menuList = append(menuList, menu)
	}

	return menuList, nil
}

//关闭一个连接
func (c *ConnectController) ActionCloseConnect(args []byte) error {
	var params struct {
		ServerId int `json:"server_id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return err
	}

	common.Conns.Remove(params.ServerId)
	return nil
}

//得到一个服务器菜单列表
func (c *ConnectController) ActionRefresh(args []byte) ([]*common.ServerMenu, error) {
	var params struct {
		ServerId int `json:"server_id"`
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	conn, err := common.Conns.Get(params.ServerId)
	if err != nil {
		return nil, err
	}

	menu, err := generateServerMenu(conn)
	if err != nil {
		return nil, err
	}
	return menu, nil
}
