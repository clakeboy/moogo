package common

var Conns *Connects

type ServerMenu struct {
	Icon     string        `json:"icon"`
	Key      string        `json:"key"`
	Text     string        `json:"text"`
	Children []*ServerMenu `json:"children"`
	Type     string        `json:"type"`
	Data     interface{}   `json:"data"`
}

//服务器连接数据
type ServerConnectData struct {
	Id            int    `storm:"id,increment" json:"id"` //主键,自增长
	Name          string `json:"name" storm:"unique"`     //服务器名称
	Address       string `json:"address"`                 //服务器IP地址
	Port          string `json:"port"`                    //服务器端口号
	IsAuth        bool   `json:"is_auth"`                 //是否验证用户名,0,1
	AuthDatabase  string `json:"auth_database"`           //验证用户数据库名
	AuthUser      string `json:"auth_user"`               //验证用户名
	AuthPassword  string `json:"auth_password"`           //验证用户密码
	IsSSH         bool   `json:"is_ssh"`                  //是否使用SSH,0,1
	SSHAddress    string `json:"ssh_address"`             //SSH服务IP地址
	SSHPort       string `json:"ssh_port"`                //SSH服务端口号
	SSHUser       string `json:"ssh_user"`                //SSH服务用户名
	SSHAuthMethod string `json:"ssh_auth_method"`         //SSH服务验证方法, password,private key
	SSHPassword   string `json:"ssh_password"`            //SSH服务密码
	SSHKeyFile    string `json:"ssh_key_file"`            //SSH密钥文件
	CreatedDate   int    `json:"created_date"`            //创建时间
}
