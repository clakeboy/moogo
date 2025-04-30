package service

import (
	"github.com/clakeboy/golib/components"
	"github.com/gin-gonic/gin"
	"moogo/common"
	"moogo/middles"
	"moogo/router"
)

type HttpServer struct {
	server  *gin.Engine
	isDebug bool
	isCross bool
	addr    string
}

func NewHttpServer(addr string, isDebug bool, isCross, isPProf bool) *HttpServer {
	server := &HttpServer{isCross: isCross, isDebug: isDebug, addr: addr}
	server.Init()
	if isPProf {
		server.StartPprof()
	}
	return server
}

func (h *HttpServer) Start() {
	wait := make(chan bool)
	go func() {
		err := h.server.Run(h.addr)
		if err != nil {
			wait <- true
		}
	}()
	<-wait
}

func (h *HttpServer) Init() {
	if h.isDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	h.server = gin.New()

	//使用中间件
	if h.isDebug {
		h.server.Use(gin.Logger(), gin.Recovery())
	} else {
		h.server.Use(middles.Logger(), middles.Recovery())
	}

	h.server.Use(middles.Cache())
	h.server.Use(middles.BoltDatabase())
	h.server.Use(middles.Cookie())
	//h.server.Use(gzip.Gzip(gzip.DefaultCompression))
	//h.server.Use(middles.Session())
	//websocket io
	h.server.GET("/socket.cio/*action", func(c *gin.Context) {
		components.Cross(c, h.isCross, c.Request.Header.Get("Origin"))
		common.SocketIO.Accept(c)
	})
	//跨域调用的OPTIONS
	h.server.OPTIONS("*action", func(c *gin.Context) {
		components.Cross(c, h.isCross, c.Request.Header.Get("Origin"))
	})

	//POST服务接收
	h.server.POST("/serv/:controller/:action", func(c *gin.Context) {
		components.Cross(c, h.isCross, c.Request.Header.Get("Origin"))
		controller := router.GetController(c.Param("controller"), c)
		components.CallAction(controller, c)
	})
	//GET服务
	h.server.GET("/serv/:controller/:action", func(c *gin.Context) {
		controller := router.GetController(c.Param("controller"), c)
		components.CallActionGet(controller, c)
	})

	//静态文件API接口
	h.server.Static("/app", "./html")
}

func (h *HttpServer) StartPprof() {
	components.InitPprof(h.server)
}
