package top

import (
	"testing"
)

func newRequest(name string) *Request {
	client := NewClient()
	client.AppKey = "12486123"
	client.SecretKey = "280a8edc3a899b8a1e4cb965732d2441"

	req := client.NewRequest(name)

	return req
}

func taobaokeItems() []Map {
	req := newRequest("taobao.taobaoke.items.get")
	req.Param("fields", "num_iid,title,nick,pic_url,price,click_url,commission,commission_rate,commission_num,commission_volume,shop_click_url,seller_credit_score,item_location,volume")
	req.Param("keyword", "nike")
	req.Param("nick", "qintb8")
	req.Param("page_size", "5")

	r, _, _ := req.All()
	return r
}

func TestErrorHandlingInvalidSignature(t *testing.T) {
	client := NewClient()
	client.AppKey = "12486123"
	client.SecretKey = "wrongsecretkey"
	req := client.NewRequest("taobao.users.get")
	req.Nicks("孙凤民")
	req.Fields("user_id", "nick", "location.city")
	_, err := req.One()
	if err.Error() != "Invalid signature" {
		t.Errorf("Error not correctly set %+v", err)
	}
}

func TestErrorHandlingErrorCode(t *testing.T) {
	req := newRequest("taobao.items.search")
	_, _, err := req.All()
	topErr, isTopErr := err.(*Error)

	if isTopErr && topErr.Code != "40" {
		t.Errorf("Wrong return err %+v", topErr)
	}
}

func TestErrorHandlingSubCode(t *testing.T) {
	req := newRequest("taobao.items.search")
	req.Fields("num_iid", "title", "price")
	_, _, err := req.All()
	topErr, isTopErr := err.(*Error)

	if isTopErr && topErr.SubCode != "isv.missing-parameter:search-none" {
		t.Errorf("Wrong return err %+v", topErr)
	}
}

func TestItemsSearch(t *testing.T) {
	req := newRequest("taobao.items.search")
	req.Fields("num_iid", "title", "price")
	req.Param("q", "格子衬衫")
	r, err := req.One()

	if err != nil {
		t.Errorf("Error returned %+v", err)
	}
	if r != nil && r["item_categories"] == nil {
		t.Errorf("didn't return categories %+v", r)
	}
	if r != nil && r["items"] == nil {
		t.Errorf("didn't return items %+v", r)
	}
}

func TestNewClient(t *testing.T) {
	items := taobaokeItems()

	req := newRequest("taobao.taobaoke.items.convert")
	req.Param("nick", "孙凤民")
	req.NumIids(items[2].ValueAsString("num_iid"))
	req.Param("fields", "click_url,num_iid,commission,commission_rate,commission_num,commission_volume")

	result, _ := req.One()

	if result["num_iid"] == nil {
		t.Errorf("result are empty %+v", result)
	}

}

func TestTaobaoUsersGet(t *testing.T) {
	req := newRequest("taobao.users.get")
	req.Nicks("孙凤民")
	req.Fields("user_id", "nick", "location.city")
	result, _ := req.One()
	if result["user_id"] == nil {
		t.Errorf("user are empty %+v", result)
	}
}

func TestTaobaoUserGet(t *testing.T) {
	req := newRequest("taobao.user.get")
	req.Param("nick", "孙凤民")
	req.Fields("user_id", "nick", "location.city")
	result, _ := req.One()

	if result["user_id"] == nil {
		t.Errorf("%+v", result)
	}
}



func TestSignature(t *testing.T) {

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
