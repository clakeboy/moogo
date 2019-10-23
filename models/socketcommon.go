package models

import "github.com/clakeboy/golib/utils"

const (
	//Socket事件列表
	SocketEventClone  = "clone_coll"
	SocketEventBackup = "backup"
	SocketEventImport = "import"
	//处理方法代码列表
	ErrorCode    = "error"
	BeginCode    = "start"
	ProcessCode  = "process"
	CompleteCode = "complete"
	CancelCode   = "cancel"
	//克隆每次处理数
	CloneNumber = 100
	//导出备份每次处理数
	ExportNumber = 500
)

//连接数据
type ConnParams struct {
	ServerId   int    `json:"server_id"`  //服务器连接ID
	Database   string `json:"database"`   //数据库
	Collection string `json:"collection"` //集合
}

//克隆传送数据
type CloneParams struct {
	*utils.JsonParse `json:"-"`
	Src              ConnParams `json:"src"`  //数据源
	Dest             ConnParams `json:"dest"` //数据目标
	Code             string     `json:"code"` //事件
}

//创建新克隆参数数据
func NewCloneParams(clone *CloneParams) *CloneParams {
	if clone == nil {
		clone = &CloneParams{}
	}
	clone.JsonParse = utils.NewJsonParse(clone)
	return clone
}

//导出传送数据
type ExportParams struct {
	*utils.JsonParse `json:"-"`
	Server           ConnParams `json:"server"`          //数据源
	Type             int        `json:"type"`            //导出数据类型 1 bson, 2 csv
	CollectionList   []string   `json:"collection_list"` //要导出的集合列表
	DestDir          string     `json:"dest_dir"`        //导出数据保存目录路径
	Code             string     `json:"code"`            //事件
}

//创建新导出参数
func NewExportParams(export ...*ExportParams) *ExportParams {
	var param *ExportParams
	if len(export) == 0 || export[0] == nil {
		param = &ExportParams{}
	} else {
		param = export[0]
	}
	param.JsonParse = utils.NewJsonParse(param)
	return param
}

//导出回复数据
type ExportResponse struct {
	Type       int    `json:"type"`       //进度类型 1 为collection,2 为数据进度
	Current    int    `json:"current"`    //当前处理数
	Total      int    `json:"total"`      //总数据
	Collection string `json:"collection"` //当前处理的数据集合
}

//导入数据参数
type ImportParams struct {
	*utils.JsonParse `json:"-"`
	Server           ConnParams `json:"server"`          //数据源
	Type             int        `json:"type"`            //导入数据类型 1 file, 2 folder
	CollectionList   []string   `json:"collection_list"` //要导入的集合列表
	Path             string     `json:"path"`            //文件或目录地址
	Code             string     `json:"code"`            //事件
	IsDrop           bool       `json:"is_drop"`         //是否删除原文档集合
}

func NewImportParams(data ...*ImportParams) *ImportParams {
	var param *ImportParams
	if len(data) == 0 || data[0] == nil {
		param = &ImportParams{}
	} else {
		param = data[0]
	}
	param.JsonParse = utils.NewJsonParse(param)
	return param
}

//SOCKET 返回数据
type SocketReturn struct {
	*utils.JsonParse `json:"-"`
	Code             string      `json:"code"`    //错误类型
	Message          string      `json:"message"` //错误说明
	Data             interface{} `json:"data"`    //传送数据
}

func NewSocketResult(code, msg string, data interface{}) *SocketReturn {
	js := &SocketReturn{
		Code:    code,
		Message: msg,
		Data:    data,
	}
	js.JsonParse = utils.NewJsonParse(js)
	return js
}

//克隆返回数据
type CloneResponse struct {
	Current int `json:"current"` //当前处理数
	Total   int `json:"total"`   //总数
}
