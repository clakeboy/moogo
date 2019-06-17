package mongo

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/clakeboy/golib/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

func TestNewDatabase(t *testing.T) {
	//cfg := &Config{
	//	Host:"127.0.0.1",
	//	Port:"33001",
	//	PoolSize:100,
	//	Auth:"admin",
	//	User:"root",
	//	Password:"WiaQ82n7B3L5Cz*2#10m",
	//}

	cfg := &Config{
		Host:     "127.0.0.1",
		Port:     "27017",
		PoolSize: 10,
	}

	db, err := NewDatabase(cfg)
	if err != nil {
		t.Error(err)
		return
	}
	err = db.Open()
	if err != nil {
		t.Error(err)
		return
	}

	coll := db.Database("center").Collection("sms_log")
	ctx := context.TODO()

	run := utils.NewExecTime()
	run.Start()
	//countOpt := options.Count()

	dataCount, err := coll.EstimatedDocumentCount(ctx)
	fmt.Println("data count:", dataCount)
	if err != nil {
		t.Error(err)
	}
	run.End(true)
	run.Start()
	var page int64 = 1
	var num int64 = 50
	findOpt := options.Find()
	findOpt.SetLimit(50)
	findOpt.SetSkip((page - 1) * num)

	cur, err := coll.Find(ctx,
		bson.M{"_id": bson.M{
			"$oid": "^189",
		}},
		findOpt,
	)

	if err != nil {
		t.Error(err)
	}

	defer cur.Close(ctx)

	var list []interface{}
	keysM := utils.M{}
	var keys []utils.M
	for cur.Next(ctx) {
		//data := bson.M{}
		val := bson.M{}
		data := bson.Raw{}
		_ = cur.Decode(&data)
		_ = cur.Decode(&val)
		elms, _ := data.Elements()
		for _, v := range elms {
			if _, ok := keysM[v.Key()]; ok {
				continue
			}
			keysM[v.Key()] = true
			keys = append(keys, utils.M{
				"key":       v.Key(),
				"type":      v.Value().Type,
				"type_name": v.Value().Type.String(),
			})
		}
		list = append(list, val)
	}
	run.End(true)
	utils.PrintAny(keys)
	utils.PrintAny(list)
}

func TestDatabase_Database(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     "27017",
		PoolSize: 100,
	}

	db, err := NewDatabase(cfg)
	if err != nil {
		t.Error(err)
		return
	}
	err = db.Open()
	if err != nil {
		t.Error(err)
		return
	}
	ctx := context.TODO()
	dba := db.Database("center")
	coll := dba.Collection("seller")

	idx := coll.Indexes()

	idxCur, err := idx.List(ctx)
	if err != nil {
		t.Error(err)
	}
	defer idxCur.Close(ctx)
	for idxCur.Next(ctx) {
		idxData := bson.M{}
		_ = idxCur.Decode(&idxData)
		utils.PrintAny(idxData)
	}
}

func TestDatabase_ListDatabase(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     "27017",
		PoolSize: 100,
	}

	db, err := NewDatabase(cfg)
	if err != nil {
		t.Error(err)
		return
	}
	err = db.Open()
	if err != nil {
		t.Error(err)
		return
	}

	ctx := context.TODO()
	dba := db.Database("center")

	cur, err := dba.ListCollections(ctx, bson.M{})
	if err != nil {
		t.Error(err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		data := bson.M{}
		_ = cur.Decode(&data)
		utils.PrintAny(data)
	}
}

func TestBsonJson(t *testing.T) {
	str := `{"_id": {"$oid":"5bc56493dd90b102815a205e"},"order_sn": "TESTygzuG7vCFs","amount": {"$numberInt":"200"},"red_proportion": {"$numberDouble":"0.2"},"business_insurance_price": {"$numberInt":"1000"},"compulsory_insurance_price": {"$numberInt":"0"},"car_tax": {"$numberInt":"0"},"insurance_price": {"$numberInt":"0"},"payment_amount": {"$numberInt":"0"},"sys_code": "CPIC","is_multi": true,"red_sn": "","is_sent": true,"send_date": {"$numberInt":"1539662997"},"send_error": "","red_list": [{"order_sn": "TESTygzuG7vCFs-1","amount": {"$numberInt":"100"},"red_sn": "RMP155DFB564946240B3C8588","is_sent": true,"send_error": "","send_date": {"$numberInt":"1539662996"}},{"order_sn": "TESTygzuG7vCFs-2","amount": {"$numberInt":"100"},"red_sn": "RMP155DFB5683C199FD3C8589","is_sent": true,"send_error": "","send_date": {"$numberInt":"1539662997"}}],"created_date": {"$numberInt":"1539662995"}}`
	data := bson.D{}
	//err := bson.Unmarshal([]byte(str),&data)
	err := bson.UnmarshalExtJSON([]byte(str), true, &data)
	if err != nil {
		t.Error(err)
	}
	utils.PrintAny(data)
	byteData, err := bson.Marshal(data)
	if err != nil {
		t.Error(err)
	}
	//fmt.Printf("%s",byteData)
	dochead := byteData[:4]
	//for i:=len(dochead)-1;i>=0;i-- {
	//	fmt.Println(dochead[i])
	//}
	length := make([]byte, 4)
	length[0], length[1], length[2], length[3] = dochead[3], dochead[2], dochead[1], dochead[0]
	doclen := utils.BytesToInt(length)
	fmt.Println(doclen, len(byteData))
}

func TestByteHex(t *testing.T) {
	str := `{"_id": {"$oid":"5bc56493dd90b102815a205e"},"order_sn": "TESTygzuG7vCFs","amount": {"$numberInt":"200"},"red_proportion": {"$numberDouble":"0.2"},"business_insurance_price": {"$numberInt":"1000"},"compulsory_insurance_price": {"$numberInt":"0"},"car_tax": {"$numberInt":"0"},"insurance_price": {"$numberInt":"0"},"payment_amount": {"$numberInt":"0"},"sys_code": "CPIC","is_multi": true,"red_sn": "","is_sent": true,"send_date": {"$numberInt":"1539662997"},"send_error": "","red_list": [{"order_sn": "TESTygzuG7vCFs-1","amount": {"$numberInt":"100"},"red_sn": "RMP155DFB564946240B3C8588","is_sent": true,"send_error": "","send_date": {"$numberInt":"1539662996"}},{"order_sn": "TESTygzuG7vCFs-2","amount": {"$numberInt":"100"},"red_sn": "RMP155DFB5683C199FD3C8589","is_sent": true,"send_error": "","send_date": {"$numberInt":"1539662997"}}],"created_date": {"$numberInt":"1539662995"}}`
	src := []byte(str)
	strByte := make([]byte, len(src)*2)
	hexStr := hex.EncodeToString(src)
	distLen := hex.Encode(strByte, src)
	strByte = strByte[:distLen]
	fmt.Println(hexStr)
	fmt.Println(len(hexStr))
	fmt.Println(strByte)
	fmt.Println(len(strByte))

	deStr, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(len(deStr))
	fmt.Println(string(deStr))

	str = `asdfasdfasdf`
}
