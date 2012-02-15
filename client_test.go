package top

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	client.AppKey = "12486123"
	client.SecretKey = "280a8edc3a899b8a1e4cb965732d2441"

	req := client.NewRequest("taobao.taobaoke.items.convert")
	req.Param("nick", "孙凤民")
	req.Param("num_iids", "14009743597")
	req.Param("fields", "click_url,num_iid,commission,commission_rate,commission_num,commission_volume")

	var result map[string]interface{}

	req.Execute(&result)

	t.Errorf("%+v", result)

	if client == nil {
		t.Errorf("New client can't be nil")
	}

}

func TestSignature(t *testing.T) {

	// http://gw.api.taobao.com/router/rest?sign=76A743E65FB43AB60B8317CB10FAAC5C&timestamp=2012-02-15+23%3A02%3A05&v=2.0&app_key=12486123&method=taobao.users.get&partner_id=top-apitools&format=json&nicks=%E5%AD%99%E5%87%A4%E6%B0%91&fields=user_id,nick,sex,buyer_credit,seller_credit,location,created,last_visit,alipay_account

	client := NewClient()
	client.AppKey = "12486123"
	client.SecretKey = "280a8edc3a899b8a1e4cb965732d2441"
	client.PartnerId = "top-apitools"

	req := client.NewRequest("taobao.users.get")
	req.Param("nicks", "孙凤民")
	req.Param("fields", "user_id,nick,sex,buyer_credit,seller_credit,location,created,last_visit,alipay_account")
	req.Param("timestamp", "2012-02-15 23:02:05")

	sign, _ := req.SignatureAndQueryString()

	if sign != "8E170B6351CC5C6CF34DFFDC63569734" {
		t.Errorf("Wrong signature %s", sign)
	}


}
