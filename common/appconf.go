package common

import (
	"context"
	"github.com/clakeboy/golib/utils"
	"os"
	"time"
)

var AppConf *AppConfig

type AppConfig struct {
	ExecTimeout int `json:"exec_timeout"` //查询超时时间 单位秒
}

func InitAppConfig() {
	AppConf = new(AppConfig)
	AppConf.ExecTimeout = 60
}

func GetAppDataDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	if !utils.Exist(dir + "/moogo") {
		_ = os.Mkdir(dir+"/moogo", 0755)
	}
	return dir + "/moogo"
}

//得到统一的连接content
func GetContent() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(AppConf.ExecTimeout)*time.Second)
	return ctx
}
