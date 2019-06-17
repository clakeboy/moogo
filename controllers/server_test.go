package controllers

import (
	"fmt"
	"github.com/clakeboy/golib/utils"
	"testing"
)

func TestServerController_ActionFind(t *testing.T) {
	data := utils.M{
		"id": 1,
	}

	res, err := utils.HttpPostJsonString("http://127.0.0.1:27317/serv/server/find", data)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
}

func TestServerController_ActionQuery(t *testing.T) {
	data := utils.M{
		"name":   "",
		"page":   1,
		"number": 10,
	}

	res, err := utils.HttpPostJsonString("http://127.0.0.1:27317/serv/server/query", data)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)

}
