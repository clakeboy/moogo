package common

import (
	"github.com/asdine/storm"
	"github.com/clakeboy/golib/ckdb"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var Conf *Config
var BDB *storm.DB

//总配置结构
type Config struct {
	System *SystemConfig       `json:"system" yaml:"system"`
	BDB    *BoltDBConfig       `json:"boltdb" yaml:"boltdb"`
	Cookie *CookieConfig       `json:"cookie" yaml:"cookie"`
	MDB    *ckdb.MongoDBConfig `json:"mdb" yaml:"mdb"`
}

type SystemConfig struct {
	Port string `json:"port" yaml:"port"`
	Ip   string `json:"ip" yaml:"ip"`
	Pid  string `json:"pid" yaml:"pid"`
}

//Cookie 配置
type CookieConfig struct {
	Path     string `json:"path" yaml:"path"`
	Domain   string `json:"domain" yaml:"domain"`
	Source   bool   `json:"source" yaml:"source"`
	HttpOnly bool   `json:"http_only" yaml:"http_only"`
}

//boltdb 配置
type BoltDBConfig struct {
	Path string `json:"path" yaml:"path"`
}

//读取一个YAML配置文件
func NewYamlConfig(confFile string) *Config {
	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		panic(err)
	}

	var conf Config
	yaml.Unmarshal(data, &conf)

	return &conf
}
