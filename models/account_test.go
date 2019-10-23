package models

import (
	"fmt"
	"github.com/clakeboy/golib/ckdb"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"moogo/common"
	"path/filepath"
	"testing"
)

func init() {
	var err error
	common.Conf = common.NewYamlConfig("../dev.conf")

	//err = ckdb.InitMongo(common.Conf.MDB)
	//if err != nil {
	//	panic(err)
	//}

	if err != nil {
		fmt.Println("open database error:", err)
	}
}

func TestNewAccountModel(t *testing.T) {
	tab := ckdb.NewDB("dudu_deduction").Collection("dudu_order")
	where := bson.M{
		"device_number": "164726775194",
		"is_paid":       0,
		"pay_date":      bson.M{"$gte": 1498608000, "$lte": 1546855577},
	}
	group, err := tab.Aggregate(bson.M{"$match": where}, bson.M{
		"$group": bson.M{
			"_id": nil,
			"count": bson.M{
				"$sum": "$ord_fee",
			},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(group["count"].(int))
}

func TestAccountModel_Query(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "moogo_export_")
	fmt.Println(tmpDir, err)
	str := "clake.bson"

	fmt.Println(filepath.Abs(str))
	fmt.Println(filepath.Ext(str))
	fmt.Println(filepath.Base(str))
	fmt.Println(filepath.Clean(str))
	fmt.Println(filepath.Dir(str))
	fmt.Println(filepath.Glob(str))
}
