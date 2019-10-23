package controllers

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/clakeboy/golib/utils"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	ReadCompress = "file"
	ReadFolder   = "folder"
)

//控制器
type BackupController struct {
	c *gin.Context
}

func NewBackupController(c *gin.Context) *BackupController {
	return &BackupController{c: c}
}

//读取导入文件列表
func (b *BackupController) ActionRead(args []byte) ([]utils.M, error) {
	var params struct {
		ReadType   string `json:"read_type"`   //读取类型,1为压缩文件,2为目录
		TargetPath string `json:"target_path"` //目标路径
	}

	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	switch params.ReadType {
	case ReadCompress:
		return b.readCompressFile(params.TargetPath)
	case ReadFolder:
		return b.readFolder(params.TargetPath)
	default:
		return nil, errors.New("not support read type")
	}
}

func (b *BackupController) readFolder(dirPath string) ([]utils.M, error) {
	list, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	var fileList []utils.M
	for _, info := range list {
		if !info.IsDir() {
			ext := filepath.Ext(info.Name())
			if ext != ".bson" {
				continue
			}

			fileList = append(fileList, utils.M{
				"collection": info.Name(),
				"size":       info.Size(),
			})
		}
	}
	return fileList, nil
}

func (b *BackupController) readCompressFile(filePath string) ([]utils.M, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	var fileList []utils.M
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ext := filepath.Ext(header.Name)
		if ext != ".bson" {
			continue
		}
		fileList = append(fileList, utils.M{
			"collection": header.Name,
			"size":       header.Size,
		})
	}
	return fileList, nil
}
