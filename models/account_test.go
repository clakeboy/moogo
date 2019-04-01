package models

import (
	"fmt"
	"github.com/clakeboy/golib/ckdb"
	"github.com/clakeboy/golib/utils"
	"gopkg.in/mgo.v2/bson"
	"moogo/common"
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
	fmt.Println(utils.RandStr(16, nil))
}
