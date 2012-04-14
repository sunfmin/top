package top

import (
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func newRequest(name string) *Request {
	client := NewClient()
	client.AppKey = "12486123"
	client.SecretKey = "280a8edc3a899b8a1e4cb965732d2441"

	req := client.NewRequest(name)

	return req
}

func taobaokeItems() []*Item {
	req := newRequest("taobao.taobaoke.items.get")
	req.Param("fields", "num_iid,title,nick,pic_url,price,click_url,commission,commission_rate,commission_num,commission_volume,shop_click_url,seller_credit_score,item_location,volume")
	req.Param("keyword", "nike")
	req.Param("nick", "qintb8")
	req.Param("page_size", "5")

	r := []*Item{}
	req.Execute(&r)
	return r
}

func TestBanSeconds(t *testing.T) {
	err := &Error{
		Code:       7,
		Message:    "App Call Limited",
		SubCode:    "accesscontrol.limited-by-app-access-count",
		SubMessage: "This ban will last for 41 more seconds",
	}
	if err.BanSeconds() != 41 {
		t.Errorf("ban seconds should be 41, but was %d", err.BanSeconds())
	}
}

func TestErrorHandlingInvalidSignature(t *testing.T) {
	client := NewClient()
	client.AppKey = "12486123"
	client.SecretKey = "wrongsecretkey"
	req := client.NewRequest("taobao.users.get")
	req.Nicks("孙凤民")
	req.Fields("user_id", "nick", "location.city")
	_, err := req.Execute(nil)
	if err.Error() != "Invalid signature" {
		t.Errorf("Error not correctly set %+v", err)
	}
}

func TestErrorHandlingErrorCode(t *testing.T) {
	req := newRequest("taobao.items.search")
	_, err := req.Execute(nil)
	topErr, isTopErr := err.(*Error)

	if isTopErr && topErr.Code != 40 {
		t.Errorf("Wrong return err %+v", topErr)
	}
}

func TestErrorHandlingSubCode(t *testing.T) {
	req := newRequest("taobao.items.search")
	req.Fields("num_iid", "title", "price")
	_, err := req.Execute(nil)
	topErr, isTopErr := err.(*Error)

	if isTopErr && topErr.SubCode != "isv.missing-parameter:search-none" {
		t.Errorf("Wrong return err %+v", topErr)
	}
}

type Item struct {
	Num_iid int64
	Title   string
	Price   string
}
type Category struct {
	Category_id int64
	Count       int64
}

type Item_categories struct {
	Item_categorie *[]Category
}

type Items struct {
	Item *[]Item
}

type ItemSearchResult struct {
	Item_categories Item_categories
	Items           Items
}

func TestItemsSearchNotStudpidWay(t *testing.T) {
	req := newRequest("taobao.items.search")
	req.Fields("num_iid", "title", "price")
	req.Param("q", "格子衬衫")

	categories := []Category{}
	items := []Item{}
	_, err := req.ExecuteIntoBranches(map[string]interface{}{
		"item_categories": &categories,
		"items":           &items,
	})

	if err != nil {
		t.Errorf("Error returned %+v", err)
	}

	if len(categories) == 0 {
		t.Errorf("didn't return categories %+v", categories)
	}

	if len(items) == 0 {
		t.Errorf("didn't return items %+v", items)
	}
}

type ItemCat struct {
	Cid        int64
	Parent_cid int64
	Name       string
	Status     string
	Sort_order int64
}

func TestItemsCatsGet(t *testing.T) {
	req := newRequest("taobao.itemcats.get")
	req.Fields("cid,parent_cid,name,is_parent")
	req.Param("cids", "50011999")

	lastModified := ""
	itemCats := []ItemCat{}
	_, err := req.ExecuteIntoBranches(map[string]interface{}{
		"last_modified": &lastModified,
		"item_cats":     &itemCats,
	})

	if err != nil {
		_, err = req.Execute(&itemCats)
	}

	if err != nil {
		t.Errorf("Error returned %+v", err)
	}

	if len(itemCats) == 0 {
		t.Errorf("should got value but empty", itemCats)
	}
}

func TestItemsSearch(t *testing.T) {
	req := newRequest("taobao.items.search")
	req.Fields("num_iid", "title", "price")
	req.Param("q", "格子衬衫")

	r := &ItemSearchResult{}

	_, err := req.Execute(r)

	if err != nil {
		t.Errorf("Error returned %+v", err)
	}
}

func TestNewClient(t *testing.T) {
	items := taobaokeItems()

	req := newRequest("taobao.taobaoke.items.convert")
	req.Param("nick", "qintb8")
	req.NumIids(items[2].Num_iid)
	req.Param("fields", "click_url,num_iid,commission,commission_rate,commission_num,commission_volume")

	newItems := []*Item{}
	req.Execute(&newItems)

	if items[0].Num_iid == 0 {
		t.Errorf("result are empty %+v", newItems)
	}

}

func TestUnmashalIntoBranches(t *testing.T) {
	f, _ := os.Open("fixtures/branches.json")
	defer f.Close()
	b, _ := ioutil.ReadAll(f)
	cleanjson, _, err := unwrapjson(b)
	if err != nil {
		t.Errorf("%+v", err)
	}

	items := []Item{}
	rmap := map[string]interface{}{
		"items": &items,
	}
	err = unmashalIntoBranches(cleanjson, rmap)

	if err != nil {
		t.Errorf("%+v", err)
	}
	if len(items) == 0 {
		t.Errorf("length should > 0, %+v", items)
	}
}

type User struct {
	User_id int64
	Nick    string
}

func TestTaobaoUsersGet(t *testing.T) {
	req := newRequest("taobao.users.get")
	req.Nicks("孙凤民")
	req.Fields("user_id", "nick", "location.city")
	us := []*User{}
	req.Execute(&us)
	if len(us) == 0 {
		t.Errorf("user are empty %+v", us)
	}
	if us[0].Nick != "孙凤民" {
		t.Errorf("user are wrong %+v", us)
	}
}

func TestTaobaoUserGet(t *testing.T) {
	req := newRequest("taobao.user.get")
	req.Param("nick", "孙凤民")
	req.Fields("user_id", "nick", "location.city")
	u := &User{}
	req.Execute(u)

	if u.Nick != "孙凤民" {
		t.Errorf("%+v", u)
	}

	if u.User_id == 0 {
		t.Errorf("%+v", u)
	}
}

func TestRequestNewSessionKey(t *testing.T) {
	client := NewClient()
	key, err := client.RequestNewSessionKey("invalid")
	if err == nil {
		t.Errorf("auth code should be invalid: %+v", key)
	}
}

func TestSessionKeyRequest(t *testing.T) {
	req := newRequest("taobao.taobaoke.report.get")
	req.Client.SessionKey = "invalid"

	req.Param("date", "20120102")
	req.Param("fields", "trade_id,pay_time,pay_price,num_iid,outer_code,commission_rate,commission,seller_nick,pay_time,app_key")
	_, err := req.Execute(nil)
	if err == nil {
		t.Errorf("session should be invalid")
	}
}

type TaobaoKeReportMember struct {
	Trade_id  int64
	Pay_price string
	Num_iid   int64
}

type Trade struct {
	Buyer_email     string
	Buyer_alipay_no string
	Buyer_nick      string
}

func TestTradeGet(t *testing.T) {
	req := newRequest("taobao.taobaoke.report.get")
	req.Client.SessionKey = "6100a16cfe79441673595a931b95b8b5288de4344cb7537322176867"
	req.Param("date", "20120225")
	req.Nicks("qintb8")
	req.Fields("trade_id", "pay_price", "num_iid")

	var r []*TaobaoKeReportMember
	req.Execute(&r)

	req2 := newRequest("taobao.trade.get")

	req2.Client.SessionKey = "6100a16cfe79441673595a931b95b8b5288de4344cb7537322176867"
	req2.Fields("buyer_email", "buyer_alipay_no", "buyer_nick")
	req2.Param("tid", strconv.FormatInt(r[0].Trade_id, 10))
	var trade Trade
	_, err := req2.Execute(&trade)
	if err == nil {
		t.Errorf("expected err to be nil but: %+v", err)
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

var unwrapjsonTests = []struct {
	jsonIn  string
	jsonOut string
	err     error
	count   int64
}{
	{`{"error_response":{"code":40,"msg":"Missing required arguments:nicks"}}`, "", &Error{40, "Missing required arguments:nicks", "", ""}, 0},
	{`{"response":{ "data": [{ "name": "Felix"}, {"name": "Juice"}], "total_results": 2}}`, `[{ "name": "Felix"}, {"name": "Juice"}]`, nil, 2},
	{`{"response":{ "data":   { "name": "Felix", "gender": "Man"}}}`, `{ "name": "Felix", "gender": "Man"}`, nil, 1},
	{`{"taobaoke_items_convert_response":{"taobaoke_items":{"taobaoke_item":[{"click_url":"http://s.click.taobao.com/t_8?e=7HZ6jHSTbIQ1Foq6TZKhbPgDOyt1AzILzvAKsVn%2B8AHubbgZmH6ogOs4aK0SlekAExSG1t9VJD3Ki0Gl7hB5WACvfhKaUqfLsduq80eGHYLvPdXowQ%3D%3D&p=mm_30129436_0_0&n=19","commission":"5.39","commission_num":"0","commission_rate":"150.00","commission_volume":"0.00","num_iid":14009743597}]},"total_results":1}}`,
		`[{"click_url":"http://s.click.taobao.com/t_8?e=7HZ6jHSTbIQ1Foq6TZKhbPgDOyt1AzILzvAKsVn%2B8AHubbgZmH6ogOs4aK0SlekAExSG1t9VJD3Ki0Gl7hB5WACvfhKaUqfLsduq80eGHYLvPdXowQ%3D%3D&p=mm_30129436_0_0&n=19","commission":"5.39","commission_num":"0","commission_rate":"150.00","commission_volume":"0.00","num_iid":14009743597}]`,
		nil,
		1,
	},
	{`{"items_search_response":{"total_results":0}}`, ``, errors.New("Unwrapped to blank"), 0},
}

func TestUnwrapjson(t *testing.T) {
	for _, tt := range unwrapjsonTests {
		cleanjson, count, err := unwrapjson([]byte(tt.jsonIn))

		stringcleanjson := string(cleanjson)

		if stringcleanjson != tt.jsonOut {
			t.Errorf("expected result is %+v, but was %+v", tt.jsonOut, stringcleanjson)
		}

		if count != tt.count {
			t.Errorf("expected count is %+v, but was %+v", tt.count, count)
		}

		if tt.err != nil && !reflect.DeepEqual(err, tt.err) {
			t.Errorf("expected err is %+v, but was %+v", tt.err, err)
		}
		if tt.err == nil && err != nil {
			t.Errorf("expected no err, but has %+v", err)
		}

	}
}
