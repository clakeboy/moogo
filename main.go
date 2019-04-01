package main

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/clakeboy/golib/utils"
	"moogo/command"
	"moogo/common"
	"moogo/service"
	"os"
	"path"
)

var out chan os.Signal
var server *service.HttpServer

func main() {
	go utils.ExitApp(out, func(s os.Signal) {
		os.Remove(command.CmdPidName)
	})
	server.Start()
}

func init() {
	var err error
	command.InitCommand()

	common.Conf = common.NewYamlConfig(command.CmdConfFile)

	if !utils.PathExists(path.Dir(common.Conf.BDB.Path)) {
		os.MkdirAll(path.Dir(common.Conf.BDB.Path), 0775)
	}
	common.BDB, err = storm.Open(common.Conf.BDB.Path)

	if err != nil {
		fmt.Println("open database error:", err)
	}

	if common.Conf.System.Pid != "" {
		command.CmdPidName = common.Conf.System.Pid
	}
	utils.WritePid(command.CmdPidName)
	out = make(chan os.Signal, 1)
	server = service.NewHttpServer(common.Conf.System.Ip+":"+common.Conf.System.Port, command.CmdDebug, command.CmdCross, command.CmdPProf)
}
