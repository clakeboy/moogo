package controllers

import (
	"fmt"
	"github.com/clakeboy/golib/utils"
	"moogo/common"
	"moogo/components/mongo"
	"moogo/models"
	"moogo/socket"
)

type SocketController struct {
	so          *socket.WebSocketClient
	cancelClone bool //是否取消克隆

}

func NewSocketController() *SocketController {
	return &SocketController{}
}

//绑定socket 事件处理器
func (s *SocketController) SocketProcess(so *socket.WebSocketClient) {
	s.so = so
	so.On(models.SocketEventBackup, s.onBackup)
	so.On(models.SocketEventClone, s.onCloneCollection)
	so.On(models.SocketEventImport, s.onImport)
}

//导入事件
func (s *SocketController) onImport(data []byte) []byte {
	params := models.NewImportParams()
	err := params.ParseJson(data)
	if err != nil {
		return models.NewSocketResult(models.ErrorCode, err.Error(), nil).ToJson()
	}
	if params.Code == models.CancelCode {
		//s.cancelClone = true
		models.CancelImport = true
		return models.NewSocketResult(models.CancelCode, "import canceled", nil).ToJson()
	}
	conn, err := common.Conns.Get(params.Server.ServerId)
	if err != nil {
		return models.NewSocketResult(models.ErrorCode, err.Error(), nil).ToJson()
	}
	dbImport := models.NewImport(params, conn, s.so)
	return dbImport.Import()
}

//备份事件
func (s *SocketController) onBackup(data []byte) []byte {
	params := models.NewExportParams()
	err := params.ParseJson(data)
	if err != nil {
		return models.NewSocketResult(models.ErrorCode, err.Error(), nil).ToJson()
	}
	if params.Code == models.CancelCode {
		//s.cancelClone = true
		models.CancelExport = true
		return models.NewSocketResult(models.CancelCode, "export canceled", nil).ToJson()
	}
	conn, err := common.Conns.Get(params.Server.ServerId)
	if err != nil {
		return models.NewSocketResult(models.ErrorCode, err.Error(), nil).ToJson()
	}
	export := models.NewExport(params, conn, s.so)
	return export.Export()
}

//克隆事件
func (s *SocketController) onCloneCollection(data []byte) []byte {
	params := models.NewCloneParams(nil)
	err := params.ParseJson(data)
	if err != nil {
		return models.NewSocketResult(models.ErrorCode, err.Error(), nil).ToJson()
	}
	fmt.Println(string(data))
	if params.Code == models.CancelCode {
		s.cancelClone = true
		return models.NewSocketResult(models.CancelCode, "clone canceled", nil).ToJson()
	}

	srcConn, err := common.Conns.Get(params.Src.ServerId)
	if err != nil {
		return models.NewSocketResult(models.ErrorCode, fmt.Sprintf("src conn err:%s", err), nil).ToJson()
	}
	destConn, err := common.Conns.Get(params.Dest.ServerId)
	if err != nil {
		return models.NewSocketResult(models.ErrorCode, fmt.Sprintf("dest conn err:%s", err), nil).ToJson()
	}

	srcColl := srcConn.Db.SelectDatabase(params.Src.Database).Collection(params.Src.Collection)
	dataCount, err := srcColl.Count(nil)
	if err != nil {
		return models.NewSocketResult(models.ErrorCode, err.Error(), nil).ToJson()
	}
	var pages int
	if dataCount%models.CloneNumber == 0 {
		pages = int(dataCount) / models.CloneNumber
	} else {
		pages = int(dataCount)/models.CloneNumber + 1
	}

	//dest collection
	destColl := destConn.Db.SelectDatabase(params.Dest.Database).Collection(params.Dest.Collection)

	go s.cloneProcess(pages, models.CloneNumber, int(dataCount), srcColl, destColl)
	return models.NewSocketResult(models.BeginCode, "ok", &models.CloneResponse{
		Current: 0,
		Total:   int(dataCount),
	}).ToJson()
}

//开始处理克隆
func (s *SocketController) cloneProcess(pages, number, dataCount int, srcColl *mongo.Collection, destColl *mongo.Collection) {
	for i := 1; i <= pages; i++ {
		if s.cancelClone {
			s.cancelClone = false
			break
		}
		list, err := srcColl.List(nil, int64(i), int64(number), nil, nil)
		if err != nil {
			_ = s.so.Emit(models.SocketEventClone, models.NewSocketResult(models.ErrorCode, err.Error(), nil).ToJson(), nil)
			break
		}
		err = destColl.Insert(list...)
		if err != nil {
			_ = s.so.Emit(models.SocketEventClone, models.NewSocketResult(models.ErrorCode, err.Error(), nil).ToJson(), nil)
			break
		}

		_ = s.so.Emit(models.SocketEventClone, models.NewSocketResult(models.ProcessCode, "processing", &models.CloneResponse{
			Current: utils.YN(i == pages, dataCount, i*number).(int),
			Total:   dataCount,
		}).ToJson(), nil)
	}
	_ = s.so.Emit(models.SocketEventClone, models.NewSocketResult(models.CompleteCode, "clone collection complete", &models.CloneResponse{
		Current: dataCount,
		Total:   dataCount,
	}).ToJson(), nil)
}
