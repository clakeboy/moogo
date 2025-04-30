package service

import (
	"moogo/controllers"
	"moogo/socket"
)

type SocketServer struct {
	engine     *socket.Engine
	controller *controllers.SocketController
}

func NewSocketServer(engine *socket.Engine) *SocketServer {
	socketServer := &SocketServer{
		engine:     engine,
		controller: controllers.NewSocketController(),
	}
	socketServer.Init()
	return socketServer
}

func (s *SocketServer) Init() {
	_ = s.engine.On(socket.EventConnect, func(so *socket.WebSocketClient) {
		s.controller.SocketProcess(so)
	})
}
