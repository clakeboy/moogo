package models

import (
	"encoding/hex"
	"fmt"
	"github.com/clakeboy/golib/utils"
	"testing"
)

func TestMath(t *testing.T) {
	s := utils.IntToBytes(100000, 32)
	fmt.Println(hex.EncodeToString(s))
	str, _ := hex.DecodeString("E68891E698AFE8BF99E4B8AA636C616B65E79A84536F636B6574494F")
	fmt.Println(string(str))
	fmt.Println(utils.RandStr(24, nil))
}

func TestPost(t *testing.T) {
	data := utils.M{
		"agencyCode":       "AC100373",
		"notifyUrl":        "https://shop.tubaozhang.com:9608/serv/shop/al_pay_async_notice",
		"payTime":          "2019-06-18 18:24:13",
		"paymentGatewaySn": "20190618182358750713",
		"paymentMethod":    "wxpay",
		"policyRef":        "301-1-593-19-0000459117-00",
		"returnUrl":        "https://shop.tubaozhang.com:9608/serv/shop/al_pay_sync_notice",
		"totalPremium":     "2377",
		"tradeStatus":      "SUCCESS",
		"sign":             "ed173cff22e3c5a08ef22cf5335d0429",
	}

	client := utils.NewHttpClient()
	res, err := client.Post("https://si.tubaozhang.com/serv/shop/al_pay_async_notice", data)
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Println(string(res))
}
