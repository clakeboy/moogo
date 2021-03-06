package models

import (
	"archive/tar"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"github.com/clakeboy/golib/utils"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"moogo/common"
	"moogo/socket"
	"os"
	"path/filepath"
	"strings"
)

const (
	ImportFile = iota + 1
	ImportFolder
)

const (
	ImportBson = iota + 1
	ImportCsv
)

var CancelImport = false

type Import struct {
	Conn              *common.Conn //导出的数据源连接
	Params            *ImportParams
	so                *socket.WebSocketClient
	TempDir           string //数据处理临时目录
	ImportLengthCache int    //导入缓存长度
}

func NewImport(params *ImportParams, conn *common.Conn, so *socket.WebSocketClient) *Import {
	return &Import{
		so:                so,
		Params:            params,
		Conn:              conn,
		ImportLengthCache: 100,
	}
}

func (i *Import) Import() []byte {
	//创建临时目录
	i.TempDir = fmt.Sprintf("%s/moogo_import/%s_%s", os.TempDir(), i.Params.Server.Database, utils.RandStr(16, nil))
	if !utils.Exist(i.TempDir) {
		err := os.MkdirAll(i.TempDir, 0755)
		if err != nil {
			return NewSocketResult(ErrorCode, err.Error(), nil).ToJson()
		}
	}

	go i.start()
	return NewSocketResult(BeginCode, "start import", nil).ToJson()
}

func (i *Import) start() {
	switch i.Params.Type {
	case ImportFile:
		i.processFile()
	case ImportFolder:
		i.processFolder()
	}
}

func (i *Import) processFile() {
	zf, err := os.Open(i.Params.Path)
	if err != nil {
		i.emitError(err.Error())
		return
	}
	defer zf.Close()
	gr, err := gzip.NewReader(zf)
	if err != nil {
		i.emitError(err.Error())
		return
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	index := 1
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			i.emitError(err.Error())
			return
		}
		isFind := false
		var currentCollection *ImportCollection
		for _, v := range i.Params.CollectionList {
			if v.Collection == header.Name {
				isFind = true
				currentCollection = v
				break
			}
		}
		//isFind, _ := utils.Contains(header.Name, i.Params.CollectionList)
		if !isFind {
			continue
		}
		collection := strings.Replace(filepath.Base(header.Name), filepath.Ext(header.Name), "", -1)
		//发送开始信息
		_ = i.so.Emit(SocketEventImport, NewSocketResult(ProcessCode, "processing", &ExportResponse{
			Type:       1,
			Collection: collection,
			Current:    index,
			Total:      len(i.Params.CollectionList),
		}).ToJson(), nil)
		switch i.Params.ImportType {
		case ImportBson:
			err = i.processBson(tr, collection)
		case ImportCsv:
			err = i.processCsv(tr, currentCollection)
		default:
			err = i.processBson(tr, collection)
		}
		//err = i.processBson(tr, collection)
		if err != nil {
			i.emitError(err.Error())
			return
		}
		index++
	}

	_ = i.so.Emit(SocketEventImport, NewSocketResult(CompleteCode, "import database complete", &ExportResponse{
		Current: len(i.Params.CollectionList),
		Total:   len(i.Params.CollectionList),
	}).ToJson(), nil)
}

func (i *Import) processFolder() {
	for idx, v := range i.Params.CollectionList {
		f, err := os.Open(fmt.Sprintf("%s/%s", i.Params.Path, v.Collection))
		if err != nil {
			i.emitError(err.Error())
			return
		}
		if v.Override != "" {
			v.Collection = v.Override
		} else {
			v.Collection = strings.Replace(v.Collection, filepath.Ext(v.Collection), "", -1)
		}

		_ = i.so.Emit(SocketEventImport, NewSocketResult(ProcessCode, "processing", &ExportResponse{
			Type:       1,
			Collection: v.Collection,
			Current:    idx + 1,
			Total:      len(i.Params.CollectionList),
		}).ToJson(), nil)
		switch i.Params.ImportType {
		case ImportBson:
			err = i.processBson(f, v.Collection)
		case ImportCsv:
			err = i.processCsv(f, v)
		default:
			err = i.processBson(f, v.Collection)
		}
		//err = i.processBson(f, collection)
		if err != nil {
			i.emitError(err.Error())
			return
		}
		f.Close()
	}
	_ = i.so.Emit(SocketEventImport, NewSocketResult(CompleteCode, "import database complete", &ExportResponse{
		Current: len(i.Params.CollectionList),
		Total:   len(i.Params.CollectionList),
	}).ToJson(), nil)
}

func (i *Import) processBson(rd io.Reader, collection string) error {
	if i.Params.IsDrop {
		err := i.Conn.Db.Database(i.Params.Server.Database).Collection(collection).Drop(nil)
		return err
	}
	coll := i.Conn.Db.Collection(collection, i.Params.Server.Database)
	var cacheList []interface{}
	dataCount := 0
	readCount := 0
	for {
		header, err := i.reRead(rd, 4)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		dataLen := i.convertBsonLength(header)
		fmt.Println(header, dataLen)
		dataByte, err := i.reRead(rd, dataLen-4)
		if err != nil {
			if err != io.EOF {
				return err
			}
			//break
		}
		readCount += dataLen
		fmt.Println(header, dataLen)
		data := bson.D{}
		err = bson.Unmarshal(append(header, dataByte...), &data)
		if err != nil {
			//i.emitError(err.Error())
			return err
		}
		cacheList = append(cacheList, data)

		if len(cacheList) >= i.ImportLengthCache {
			err = coll.Insert(cacheList...)
			if err != nil {
				return err
			}
			dataCount += len(cacheList)
			_ = i.so.Emit(SocketEventImport, NewSocketResult(ProcessCode, "processing", &ExportResponse{
				Type:       2,
				Collection: collection,
				Current:    dataCount,
				Total:      readCount,
			}).ToJson(), nil)
			cacheList = nil
		}
	}
	if len(cacheList) > 0 {
		err := coll.Insert(cacheList...)
		if err != nil {
			return err
		}
		dataCount += len(cacheList)
		_ = i.so.Emit(SocketEventImport, NewSocketResult(ProcessCode, "processing", &ExportResponse{
			Type:       2,
			Collection: collection,
			Current:    dataCount,
			Total:      readCount,
		}).ToJson(), nil)
	}
	return nil
}

func (i *Import) reRead(rd io.Reader, readCount int) ([]byte, error) {
	cache := make([]byte, readCount)
	rn, err := rd.Read(cache)
	if err != nil {
		return cache, err
	}
	if rn != readCount {
		tmp, err := i.reRead(rd, readCount-rn)
		if err != nil {
			return nil, err
		}
		cache = append(cache[:rn], tmp...)
	}
	return cache, nil
}

func (i *Import) convertBsonLength(header []byte) int {
	tmp := make([]byte, 4)
	tmp[0], tmp[1], tmp[2], tmp[3] = header[3], header[2], header[1], header[0]
	return utils.BytesToInt(tmp)
}

func (i *Import) emitError(msg string) {
	_ = i.so.Emit(SocketEventImport, NewSocketResult(ErrorCode, msg, nil).ToJson(), nil)
}

func (i *Import) processCsv(rd io.Reader, collection *ImportCollection) error {
	if i.Params.IsDrop {
		err := i.Conn.Db.Database(i.Params.Server.Database).Collection(collection.Collection).Drop(nil)
		return err
	}
	coll := i.Conn.Db.Collection(collection.Collection, i.Params.Server.Database)
	var cacheList []interface{}
	dataCount := 0
	readCount := 0
	csvRd := csv.NewReader(rd)
	var header []string
	var err error
	if collection.IsFirstColumn {
		header, err = csvRd.Read()
		if err != nil {
			return err
		}
	}
	for {
		row, err := csvRd.Read()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		data := bson.D{}
		for i, v := range header {
			data = append(data, bson.E{Key: v, Value: row[i]})
		}
		cacheList = append(cacheList, data)
		if len(cacheList) >= i.ImportLengthCache {
			err = coll.Insert(cacheList...)
			if err != nil {
				return err
			}
			dataCount += len(cacheList)
			_ = i.so.Emit(SocketEventImport, NewSocketResult(ProcessCode, "processing", &ExportResponse{
				Type:       2,
				Collection: collection.Collection,
				Current:    dataCount,
				Total:      readCount,
			}).ToJson(), nil)
			cacheList = nil
		}
	}
	if len(cacheList) > 0 {
		err := coll.Insert(cacheList...)
		if err != nil {
			return err
		}
		dataCount += len(cacheList)
		_ = i.so.Emit(SocketEventImport, NewSocketResult(ProcessCode, "processing", &ExportResponse{
			Type:       2,
			Collection: collection.Collection,
			Current:    dataCount,
			Total:      readCount,
		}).ToJson(), nil)
	}
	return nil
}
