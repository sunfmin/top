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
	req.Param("num_iids", "13342448437")
	req.Param("fields", "click_url,num_iid,commission,commission_rate,commission_num,commission_volume")

	var result map[string]interface{}

	req.Execute(&result)

	if client == nil {
		t.Errorf("New client can't be nil")
	}

}

