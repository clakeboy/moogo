package command

/**
---  -> 0   (no execute , no write ,no read)
--x  -> 1   execute, (no write, no read)
-w-  -> 2   write
-wx  -> 3   write, execute
r--  -> 4   read
r-x  -> 5   read, execute
rw-  -> 6   read, write ,
rwx  -> 7   read, write , execute
*/

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/clakeboy/golib/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"moogo/common"
	"moogo/models"
	"os"
	"path"
	"time"
)

//初始化程序
func InitApp() {
	defer func() {
		Exit("init application done!")
	}()

	if !utils.Exist(CmdConfFile) {
		generateConfig()
	}

	var err error
	common.Conf = common.NewYamlConfig(CmdConfFile)

	if !utils.PathExists(path.Dir(common.Conf.BDB.Path)) {
		os.MkdirAll(path.Dir(common.Conf.BDB.Path), 0775)
	}
	common.BDB, err = storm.Open(common.Conf.BDB.Path)

	if err != nil {
		Exit(fmt.Sprintf("%v", err))
	}
	initUser()
}

//初始化默认用户
func initUser() {
	model := models.NewAccountModel(common.BDB)
	data := &models.AccountData{
		Account:     "admin",
		UserName:    "admin",
		Password:    utils.EncodeMD5("123123"),
		Disable:     false,
		CreatedDate: int(time.Now().Unix()),
	}
	err := model.Save(data)
	if err != nil {
		Exit(fmt.Sprintf("init user error :%v", err))
	}
}

//生成默认配置文件
func generateConfig() {
	conf := &common.Config{
		System: &common.SystemConfig{
			Port: "27317",
			Ip:   "",
			Pid:  "moogo.pid",
		},
		BDB: &common.BoltDBConfig{
			Path: "./db/moogo.db",
		},
		Cookie: &common.CookieConfig{
			Path:     "/",
			Domain:   "",
			Source:   false,
			HttpOnly: false,
		},
	}

	out, err := yaml.Marshal(conf)
	if err != nil {
		Exit(fmt.Sprintf("generate config file error : %v", err))
	}

	ioutil.WriteFile(CmdConfFile, out, 0644)
}
