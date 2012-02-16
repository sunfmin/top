package top

import (
	"encoding/json"
	"time"
	"sort"
	"crypto/md5"
	"io"
	"net/url"
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
	"reflect"
)

type Client struct {
	AppKey string
	SecretKey string
	PartnerId string

	signMethod string
	format  string
	v string
}

type Request struct {
	Name string
	Params map[string]string

	client *Client
}


func NewClient() *Client {
	c := &Client{}
	c.PartnerId = "top-sdk-go-20120214"
	c.format = "json"
	c.signMethod = "md5"
	c.v = "2.0"
	return c
}

func (client *Client) NewRequest(name string) *Request {
	return &Request{Name: name, Params: map[string]string{}, client: client}
}

func (req *Request) Execute() (r []map[string]interface{}, count int64, err error) {
	_, query := req.SignatureAndQueryString()

	url := "http://gw.api.taobao.com/router/rest?" + query
	fmt.Printf("Requesting: %+v\n\n", url)

	resp, err := http.Get(url)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
/*	fmt.Println(string(body))*/
	var result interface{}
	json.Unmarshal(body, &result)
	tm := &taobaoMap{result, 1}
	unwrapped := tm.unwrap()
	count = tm.count
	r = unwrapped.result()
	return
}

func (req *Request) One() (r map[string]interface{}, err error) {
	res, _, err := req.Execute()
	r = res[0]
	return
}

func (req *Request) All() (r []map[string]interface{}, count int64, err error) {
	r, count, err = req.Execute()
	return
}


func (req *Request) Param(name string, value string) {
	req.Params[name] = value
}

func (req *Request) SignatureAndQueryString() (sign string, qs string) {
	ps := req.makeRequestParams()

	var keys []string

	for k, _ := range(ps) {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	keyvalues := "";

	for _, k := range(keys) {
		keyvalues = keyvalues + k + ps[k]
	}

    beforeMd5 := req.client.SecretKey + keyvalues + req.client.SecretKey

	c := md5.New()
	io.WriteString(c, beforeMd5)
	sign = fmt.Sprintf("%X", c.Sum(nil))

	values := &url.Values{}
	values.Add("sign", sign)
	for _, k := range(keys) {
		values.Add(k, ps[k])
	}
	qs = values.Encode()
	return
}

func (req *Request) Fields(fields ...string) {
	req.Params["fields"] = strings.Join(fields, ",")
}

func (req *Request) NumIids(ids ...interface{}) {
	var strIds []string
	for _, v := range(ids) {
		switch newval := v.(type) {
		case float64:
			strIds = append(strIds, strconv.FormatFloat(newval, 'f', 0, 64))
		case string:
			strIds = append(strIds, newval)
		default:
			panic(fmt.Sprintf("type not allowed %+v", reflect.TypeOf(v)))
		}
	}
	req.Params["num_iids"] = strings.Join(strIds, ",")
}

func (req *Request) Nicks(nicks ...string) {
	req.Params["nicks"] = strings.Join(nicks, ",")
}

func (req *Request) makeRequestParams() map[string]string {
	ps := map[string]string{}
	ps["v"] = req.client.v
	ps["sign_method"] = req.client.signMethod
	ps["app_key"] = req.client.AppKey
	ps["partner_id"] = req.client.PartnerId
	ps["method"] = req.Name
	ps["format"] = req.client.format
	ps["timestamp"] = time.Now().Format("2006-01-02 15:04:05")

	for k, v := range(req.Params) {
		ps[k] = v
	}
	return ps
}

