package main

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/clakeboy/golib/utils"
	"moogo/command"
	"moogo/common"
	"moogo/service"
	"moogo/socket"
	"os"
)

var out chan os.Signal
var server *service.HttpServer
var socketServer *service.SocketServer

func main() {
	common.InitAppConfig()
	go utils.ExitApp(out, func(s os.Signal) {
		_ = os.Remove(command.CmdPidName)
	})
	server.Start()
}

func init() {
	var err error
	command.InitCommand()

	common.Conf = common.NewYamlConfig(command.CmdConfFile)

	common.BDB, err = storm.Open(common.GetAppDataDir() + "/moogo.s")

	if err != nil {
		fmt.Println("open database error:", err)
	}

	common.Conns = common.NewConnects()

	if common.Conf.System.Pid != "" {
		command.CmdPidName = common.Conf.System.Pid
	}

	common.SocketIO = socket.NewEngine()
	socketServer = service.NewSocketServer(common.SocketIO)

	utils.WritePid(command.CmdPidName)
	out = make(chan os.Signal, 1)
	server = service.NewHttpServer(common.Conf.System.Ip+":"+common.Conf.System.Port, command.CmdDebug, command.CmdCross, command.CmdPProf)
}
