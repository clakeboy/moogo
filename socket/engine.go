package socket

import (
	"errors"
	"github.com/clakeboy/golib/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024

	//engine event
	EventConnect    = "connection"
	EventDisconnect = "disconnection"

	//engine protocol version
	CIO = "1"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

//网络引擎
type Engine struct {
	conns  map[string]*WebSocketClient
	events map[string]*caller
	aes    *utils.AesEncrypt
	aesKey string
	rooms  map[string]map[string]*WebSocketClient
	evlk   sync.RWMutex
	rmslk  sync.RWMutex
	rmlk   sync.RWMutex
}

//创建一个新网络引擎
func NewEngine() *Engine {
	return &Engine{
		conns:  make(map[string]*WebSocketClient),
		events: make(map[string]*caller),
		rooms:  make(map[string]map[string]*WebSocketClient),
	}
}

//接受一个新的SOCKET连接
func (e *Engine) Accept(c *gin.Context) {
	protocolV := c.Query("CIO")
	protocol := c.Query("protocol")
	if protocolV != CIO {
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("not support this version"))
		return
	}

	if protocol == "transport" {
		s_key := c.Query("s")
		c.String(http.StatusOK, string(s_key))
		return
	}

	if protocol != "websocket" {
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("not support this protocol"))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := NewWebsocket(conn, e)
	client.Start()
	e.execEvent(EventConnect, client)
	e.evlk.Lock()
	e.conns[client.Id()] = client
	e.evlk.Unlock()
}

//连接离开房间
func (e *Engine) Leave(roomId string, so *WebSocketClient) {
	e.rmslk.RLock()
	if room, ok := e.rooms[roomId]; ok {
		e.rmslk.RUnlock()
		e.rmlk.Lock()
		delete(room, so.Id())
		e.rmlk.Unlock()
	} else {
		e.rmslk.RUnlock()
	}
}

//连接加入房间
func (e *Engine) Join(roomId string, so *WebSocketClient) map[string]*WebSocketClient {
	e.rmslk.RLock()
	if room, ok := e.rooms[roomId]; ok {
		e.rmslk.RUnlock()
		e.rmlk.Lock()
		room[so.Id()] = so
		e.rmlk.Unlock()
		return room
	} else {
		e.rmslk.RUnlock()
		room := make(map[string]*WebSocketClient)
		room[so.Id()] = so
		e.rmslk.Lock()
		e.rooms[roomId] = room
		e.rmslk.Unlock()
		return room
	}
}

//处理引擎事件
func (e *Engine) On(evt string, fn interface{}) error {
	call, err := newCaller(fn)
	if err != nil {
		return err
	}
	e.events[evt] = call
	return nil
}

//执行事件
func (e *Engine) execEvent(evtName string, so *WebSocketClient) {
	if call, ok := e.events[evtName]; ok {
		call.Call(so, nil)
	}
}

//引发一个引擎事件
func (e *Engine) Emit(evtStr string) {

}

//发送一个广播信息
func (e *Engine) BroadcastTo(roomId string, sid string, evtName string, data []byte) {
	e.rmslk.RLock()
	if room, ok := e.rooms[roomId]; ok {
		e.rmslk.RUnlock()
		e.rmlk.RLock()
		for _, so := range room {
			if so.Id() != sid {
				_ = so.Emit(evtName, data, nil)
			}
		}
		e.rmlk.RUnlock()
	} else {
		e.rmslk.RUnlock()
	}
}

//关闭一个连接
func (e *Engine) CloseConnect(sid string) {
	e.evlk.RLock()
	if conn, ok := e.conns[sid]; ok {
		e.evlk.RUnlock()
		e.evlk.Lock()
		delete(e.conns, sid)
		e.evlk.Unlock()
		e.execEvent(EventDisconnect, conn)
	} else {
		e.evlk.RUnlock()
	}
}

//返回连接数
func (e *Engine) Count() int {
	return len(e.conns)
}
