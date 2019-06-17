package common

import (
	"errors"
	"moogo/components/mongo"
)

type Conn struct {
	Db         *mongo.Database
	ServerInfo *ServerConnectData
	SSH        *SSHSession
}

type Connects struct {
	list map[int]*Conn
}

func NewConnects() *Connects {
	return &Connects{
		list: make(map[int]*Conn),
	}
}

func (c *Connects) Add(conn *Conn) {
	c.list[conn.ServerInfo.Id] = conn
}

func (c *Connects) Get(serverId int) (*Conn, error) {
	if conn, ok := c.list[serverId]; ok {
		return conn, nil
	}
	return nil, errors.New("the connection is disconnect")
}

func (c *Connects) Remove(serverId int) {
	conn, ok := c.list[serverId]
	if !ok {
		return
	}
	if conn.SSH != nil {
		conn.SSH.Close()
	}
	delete(c.list, serverId)
}

func (c *Connects) Len() int {
	return len(c.list)
}

func (c *Connects) Each(fn func(*Conn)) {
	for _, v := range c.list {
		fn(v)
	}
}

func (c *Connects) List() []*Conn {
	var list []*Conn
	for _, v := range c.list {
		list = append(list, v)
	}
	return list
}

func (c *Connects) HasServer(serverId int) bool {
	_, ok := c.list[serverId]
	return ok
}
