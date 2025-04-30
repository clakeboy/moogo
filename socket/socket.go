package socket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/clakeboy/golib/utils"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

const (
	NormalMessage = 0x01
	AckMessage    = 0x0f
)

const (
	SocketStatusOpen = iota + 1
	SocketStatusClose
)

//接收事件模型
type EventMessage struct {
	IsAck bool   `json:"is_ack"`
	Event string `json:"event"`
	Data  []byte `json:"data"`
}

func (e *EventMessage) ToJson() []byte {
	str, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return str
}

//scoket连接类
type WebSocketClient struct {
	ena    *Engine
	conn   *websocket.Conn
	id     string
	send   chan []byte
	close  chan bool
	events map[string]*caller
	acks   map[string]*caller
	rooms  map[string]bool
	rmlk   sync.RWMutex
	status int //连接状态
}

//创建一个新的连接类
func NewWebsocket(conn *websocket.Conn, ena *Engine) *WebSocketClient {
	return &WebSocketClient{
		ena:    ena,
		conn:   conn,
		id:     utils.RandStr(10, nil),
		events: make(map[string]*caller),
		acks:   make(map[string]*caller),
		send:   make(chan []byte),
		close:  make(chan bool, 1),
		rooms:  make(map[string]bool),
	}
}

//开始监听并处理SOCKET读写事件
func (w *WebSocketClient) Start() {
	go w.Read()
	go w.Write()
	w.status = SocketStatusOpen
}

//得到连接session ID
func (w *WebSocketClient) Id() string {
	return w.id
}

//读取连接
func (w *WebSocketClient) Read() {
	defer func() {
		fmt.Println(w.id, " close read")
		w.close <- true
		w.Close()
	}()
	w.conn.SetReadLimit(maxMessageSize)
	_ = w.conn.SetReadDeadline(time.Now().Add(pongWait))
	w.conn.SetPongHandler(func(string) error { _ = w.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		messageType, message, err := w.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("host error: %v", err)
			}
			break
		}
		switch messageType {
		case websocket.TextMessage:
			w.Receive(message)
		case websocket.BinaryMessage:
			w.Receive(message)
		case websocket.CloseMessage:
			break
		}
	}
}

//写入连接
func (w *WebSocketClient) Write() {
	defer func() {
		fmt.Println(w.id, " close write")
		close(w.send)
		close(w.close)
		w.Close()
	}()
	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case message, ok := <-w.send:
			_ = w.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				_ = w.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			nw, err := w.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			_, _ = nw.Write(message)
			// Add queued chat messages to the current websocket message.
			//n := len(c.send)
			//for i := 0; i < n; i++ {
			//	nw.Write([]byte("|||"))
			//	nw.Write([]byte(<-c.send))
			//}

			if err := nw.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = w.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := w.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-w.close:
			return
		}
	}
}

//接收事件消息并执行事件
func (w *WebSocketClient) Receive(msg []byte) {
	var deMsg []byte
	var err error
	if w.ena.aes != nil {
		deMsg, err = w.ena.aes.Decrypt(msg)
		if err != nil {
			log.Fatal(err)
			return
		}
	} else {
		deMsg = msg
	}
	evt, err := w.deProtocol(deMsg)
	if err != nil {
		log.Fatal(err)
		return
	}
	w.execEvent(evt.IsAck, evt.Event, evt.Data)
}

//下发事件消息
func (w *WebSocketClient) Emit(evtName string, data []byte, ack func()) error {
	//var c *caller
	//if l := len(args); l > 0 {
	//	fv := reflect.ValueOf(args[l-1])
	//	if fv.Kind() == reflect.Func {
	//		var err error
	//		c, err = newCaller(args[l-1])
	//		if err != nil {
	//			return err
	//		}
	//		args = args[:l-1]
	//	}
	//}
	if w.status == SocketStatusClose {
		return nil
	}
	if ack != nil {
		c, err := newCaller(ack)
		if err != nil {
			return err
		}
		w.acks[evtName] = c
	}
	w.send <- w.enProtocol(NormalMessage, evtName, data)
	return nil
}

//下发ACK消息
func (w *WebSocketClient) EmitAck(evtName string, data []byte) {
	if w.status == SocketStatusClose {
		return
	}
	w.send <- w.enProtocol(AckMessage, evtName, data)
}

//打包协议内容
func (w *WebSocketClient) enProtocol(protocol byte, evtName string, data []byte) []byte {
	evt := []byte(evtName)
	evtLength := utils.IntToBytes(len(evt), 8)
	var buf bytes.Buffer
	buf.WriteByte(protocol)
	buf.Write(evtLength)
	buf.Write(evt)
	buf.Write(data)
	return buf.Bytes()
}

//解开协议内容
func (w *WebSocketClient) deProtocol(msg []byte) (*EventMessage, error) {
	protocol := msg[0]
	evtLength := utils.BytesToInt(msg[1:2])
	evtName := string(msg[2 : 2+evtLength])
	content := msg[2+evtLength:]
	return &EventMessage{
		IsAck: protocol == AckMessage,
		Event: evtName,
		Data:  content,
	}, nil
}

//执行事件函数，完成后如有返回值就触发 ACKS 返回事件
func (w *WebSocketClient) execEvent(isAck bool, evtName string, args ...interface{}) {
	if isAck {
		if ack, ok := w.acks[evtName]; ok {
			ack.Call(w, args)
			delete(w.acks, evtName)
		}
		return
	}

	if evt, ok := w.events[evtName]; ok {
		rel := evt.Call(w, args)
		if len(rel) > 0 {
			w.EmitAck(evtName, rel[0].Bytes())
		}
	}
}

//绑定事件方法
func (w *WebSocketClient) On(evtName string, fv interface{}) {
	call, err := newCaller(fv)
	if err == nil {
		w.events[evtName] = call
	}
}

func (w *WebSocketClient) Join(roomName string) {
	w.ena.Join(roomName, w)
	w.rmlk.Lock()
	w.rooms[roomName] = true
	w.rmlk.Unlock()
}

func (w *WebSocketClient) Leave(roomId string) {
	w.rmlk.RLock()
	if _, ok := w.rooms[roomId]; ok {
		w.rmlk.Unlock()
		w.rmlk.Lock()
		delete(w.rooms, roomId)
		w.rmlk.Unlock()
		w.ena.Leave(roomId, w)
	} else {
		w.rmlk.Unlock()
	}
}

func (w *WebSocketClient) LeaveAll() {
	w.rmlk.Lock()
	defer w.rmlk.Unlock()
	for roomId := range w.rooms {
		delete(w.rooms, roomId)
		w.ena.Leave(roomId, w)
	}
}

//发送一个广播信息
func (w *WebSocketClient) BroadcastTo(roomId string, evtName string, data []byte) {
	w.ena.BroadcastTo(roomId, w.id, evtName, data)
}

//关闭socket 连接
func (w *WebSocketClient) Close() {
	_ = w.conn.Close()
	w.status = SocketStatusClose
	w.LeaveAll()
	w.ena.CloseConnect(w.id)
}
