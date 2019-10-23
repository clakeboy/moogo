package models

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/clakeboy/golib/utils"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"moogo/common"
	"moogo/socket"
	"os"
	"time"
)

const (
	ExportTypeBson = iota + 1
	ExportTypeCsv
)

var CancelExport bool = false

type Export struct {
	Params  *ExportParams           //导出参数
	Conn    *common.Conn            //导出的数据源连接
	TempDir string                  //临时目录
	so      *socket.WebSocketClient //socket 连接
}

func NewExport(params *ExportParams, conn *common.Conn, so *socket.WebSocketClient) *Export {
	return &Export{
		Params: params,
		Conn:   conn,
		so:     so,
	}
}

func (e *Export) Export() []byte {
	//创建临时目录
	tmpDir := fmt.Sprintf("%s/moogo_export/%s_%s", os.TempDir(), e.Params.Server.Database, utils.RandStr(16, nil))
	if !utils.Exist(tmpDir) {
		err := os.MkdirAll(tmpDir, 0755)
		if err != nil {
			return NewSocketResult(ErrorCode, err.Error(), nil).ToJson()
		}
	}
	e.TempDir = tmpDir
	go e.start()
	return NewSocketResult(BeginCode, "start export", nil).ToJson()
}

func (e *Export) start() {
	exportName := e.newExportName()
	fullPathName := fmt.Sprintf("%s/%s", e.Params.DestDir, exportName)
	//如果存在已导出的文件,先清除
	if utils.Exist(fullPathName) {
		err := os.Remove(fullPathName)
		if err != nil {
			e.emitError(err.Error())
			return
		}
	}
	zipFile, err := os.Create(fullPathName)
	if err != nil {
		e.emitError(err.Error())
		return
	}
	zipWriter, _ := gzip.NewWriterLevel(zipFile, 9)
	tarWriter := tar.NewWriter(zipWriter)
	defer func() {
		_ = tarWriter.Close()
		_ = zipWriter.Close()
		_ = zipFile.Close()
		if CancelExport {
			CancelExport = false
			_ = os.Remove(fullPathName)
		}
		_ = os.RemoveAll(e.TempDir)
	}()
	for i, v := range e.Params.CollectionList {
		if CancelExport {
			return
		}
		_ = e.so.Emit(SocketEventBackup, NewSocketResult(ProcessCode, "processing", &ExportResponse{
			Type:       1,
			Collection: v,
			Current:    i + 1,
			Total:      len(e.Params.CollectionList),
		}).ToJson(), nil)

		//wf, err := zipWriter.Create(fmt.Sprintf("%s/%s.%s", e.Params.Server.Database, v,e.getExportExt()))
		//if err != nil {
		//	e.emitError(err.Error())
		//	return
		//}

		err = e.process(tarWriter, v)
		if err != nil {
			e.emitError(err.Error())
			return
		}
	}
	_ = zipWriter.Flush()
	_ = e.so.Emit(SocketEventBackup, NewSocketResult(CompleteCode, "export database complete", &ExportResponse{
		Current: len(e.Params.CollectionList),
		Total:   len(e.Params.CollectionList),
	}).ToJson(), nil)
}

func (e *Export) process(tw *tar.Writer, collection string) error {
	coll := e.Conn.Db.SelectDatabase(e.Params.Server.Database).Collection(collection)
	dataCount, err := coll.Count(nil)
	if err != nil {
		return err
	}
	var pages int
	if dataCount%ExportNumber == 0 {
		pages = int(dataCount) / ExportNumber
	} else {
		pages = int(dataCount)/ExportNumber + 1
	}

	//临时写入文件
	tmpFile, err := os.Create(fmt.Sprintf(
		"%s/%s.%s",
		e.TempDir,
		collection,
		e.getExportExt(),
	))
	defer func() {
		_ = tmpFile.Close()
	}()
	for i := 1; i <= pages; i++ {
		if CancelExport {
			return nil
		}
		list, err := coll.List(nil, int64(i), int64(ExportNumber), nil, nil)
		if err != nil {
			return err
		}
		_ = e.so.Emit(SocketEventBackup, NewSocketResult(ProcessCode, "processing", &ExportResponse{
			Type:       2,
			Collection: collection,
			Current:    utils.YN(i == pages, int(dataCount), i*ExportNumber).(int),
			Total:      int(dataCount),
		}).ToJson(), nil)
		err = e.convertBson(list, tmpFile)
		if err != nil {
			return err
		}
	}
	err = tmpFile.Sync()
	if err != nil {
		return err
	}
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return err
	}
	//tar 归档写入文件信息
	info, err := tmpFile.Stat()
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = fmt.Sprintf("%s/%s.%s", e.Params.Server.Database, collection, e.getExportExt())

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, tmpFile)
	if err != nil {
		return err
	}

	return nil
}

func (e *Export) convertBson(dataList []interface{}, wf io.Writer) error {
	for _, v := range dataList {
		out, err := bson.Marshal(v)
		if err != nil {
			return err
		}
		_, err = wf.Write(out)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Export) emitError(msg string) {
	_ = e.so.Emit(SocketEventBackup, NewSocketResult(ErrorCode, msg, nil).ToJson(), nil)
}

func (e *Export) newExportName() string {
	datetime := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%s.tgz",
		e.Params.Server.Database,
		datetime,
	)
}

func (e *Export) getExportExt() string {
	switch e.Params.Type {
	case ExportTypeBson:
		return "bson"
	case ExportTypeCsv:
		return "csv"
	default:
		return ""
	}
}
