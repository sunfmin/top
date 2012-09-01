package top

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var Verbose bool

type Client struct {
	AppKey     string
	SecretKey  string
	SessionKey string
	PartnerId  string

	signMethod          string
	format              string
	v                   string
	appKeyCallCountList KeyCallCounts
	currentUsingAppKey  *appKeyCallCount
}

type Request struct {
	Name   string
	Params map[string]string

	Client *Client
}

var bansecondsRegexp = regexp.MustCompile(`This ban[^\d]*(\d+)[^\d]*seconds`)

type Error struct {
	Code       int64  `json:"code"`
	Message    string `json:"msg"`
	SubCode    string `json:"sub_code"`
	SubMessage string `json:"sub_msg"`
}

func (err *Error) BanSeconds() (r int64) {
	if err.Code != 7 {
		return 0
	}

	rs := bansecondsRegexp.FindStringSubmatch(err.SubMessage)
	r, _ = strconv.ParseInt(rs[1], 10, 64)
	return
}

func (err *Error) Error() string {
	if len(err.SubMessage) > 0 {
		return err.Message + ": " + err.SubMessage
	}
	return err.Message
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
	return &Request{Name: name, Params: map[string]string{}, Client: client}
}

func (client *Client) RequestNewSessionKey(authcode string) (sessKey string, err error) {
	hc := &http.Client{
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			sessKey = req.URL.Query().Get("top_session")
			return errors.New("found")
		}}

	r, err := hc.Get(fmt.Sprintf("http://container.open.taobao.com/container?authcode=%s&encode=utf-8", authcode))

	if sessKey != "" {
		return sessKey, nil
	}

	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	bodyBytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return "", err
	}

	vals, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		return "", err
	}

	if vals.Get("error") != "" {
		errorCode, _ := strconv.ParseInt(vals.Get("error"), 10, 64)
		return "", &Error{errorCode, vals.Get("error_description"), "", ""}
	}

	return "", errors.New("no session key")
}

func (req *Request) Execute(r interface{}) (count int64, err error) {
Retry:
	count = 0
	body, err := req.doRequestAndGetBody()
	if err != nil {
		return
	}
	cleanjson, count, err := unwrapjson(body)

	if err != nil {
		var switched bool
		err, switched = req.Client.switchedKeyIfBanned(err)
		if switched {
			goto Retry
		}
		return
	}

	err = json.Unmarshal(cleanjson, &r)
	if err != nil {
		if Verbose {
			log.Printf("top: cannot unmarshal json in top.Execute: %+v\n", string(cleanjson))
		}
		return
	}

	return
}

func (req *Request) ExecuteIntoBranches(rmap map[string]interface{}) (count int64, err error) {
Retry:
	body, err := req.doRequestAndGetBody()
	if err != nil {
		return
	}

	cleanjson, count, err := unwrapjson(body)

	if err != nil {
		var switched bool
		err, switched = req.Client.switchedKeyIfBanned(err)
		if switched {
			goto Retry
		}
		return
	}

	err = unmashalIntoBranches(cleanjson, rmap)
	if err != nil {
		if Verbose {
			log.Printf("top: cannot unmarshal json in top.ExecuteIntoBranches: %+v\n", err)
		}
		return
	}

	return
}

func (req *Request) doRequestAndGetBody() (body []byte, err error) {
	req.Client.countOrSwitchOrWait()
	_, query := req.SignatureAndQueryString()

	url := "http://gw.api.taobao.com/router/rest?" + query
	if Verbose {
		log.Printf("top: requesting: %+v\n\n", url)
	}
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	body = []byte(strings.Replace(string(body), "\r", "", -1))
	//log.Println(string(body))
	return
}

func unmashalIntoBranches(data []byte, rmap map[string]interface{}) (err error) {
	var unwraped map[string]json.RawMessage
	json.Unmarshal(data, &unwraped)

	noValueKeys := []string{}

	for name, rval := range rmap {
		unwrapedBranchValue, haveVal := unwraped[name]
		if !haveVal {
			noValueKeys = append(noValueKeys, name)
			continue
		}

		clean, _, err := unwrapjson(unwrapedBranchValue)
		if err != nil {
			return err
		}

		err = json.Unmarshal(clean, &rval)
		if err != nil {
			return err
		}
	}

	if len(noValueKeys) == len(rmap) {
		return errors.New(fmt.Sprintf("All keys %+v of branches not exist in responsed json, %+v", noValueKeys, string(data)))
	}
	return
}

func unwrapjson(data []byte) (cleanjson []byte, count int64, err error) {

	valInLoop := data
	count = 0
	for {
		var val map[string]json.RawMessage
		json.Unmarshal(valInLoop, &val)

		if len(val) == 0 {
			cleanjson = valInLoop
			break
		}

		countmsg, countExist := val["total_results"]
		if countExist {
			count, _ = strconv.ParseInt(string(countmsg), 10, 64)
			delete(val, "total_results")
		}

		if len(val) == 0 {
			cleanjson = []byte{}
			err = errors.New("Unwrapped to blank")
			break
		}

		errmsg, errExist := val["error_response"]
		if errExist {
			var topError Error
			json.Unmarshal(errmsg, &topError)
			err = error(&topError)
			break
		}

		if len(val) == 1 {
			for _, rawJson := range val {
				valInLoop = rawJson
				break
			}
		}
		if len(val) > 1 {
			if count == 0 {
				count = 1
			}
			cleanjson = valInLoop
			break
		}
	}
	return
}

func (req *Request) Param(name string, value string) {
	req.Params[name] = value
}

func (req *Request) SignatureAndQueryString() (sign string, qs string) {
	ps := req.makeRequestParams()

	var keys []string

	for k, _ := range ps {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	keyvalues := ""

	for _, k := range keys {
		keyvalues = keyvalues + k + ps[k]
	}

	beforeMd5 := req.Client.SecretKey + keyvalues + req.Client.SecretKey

	c := md5.New()
	io.WriteString(c, beforeMd5)
	sign = fmt.Sprintf("%X", c.Sum(nil))

	values := &url.Values{}
	values.Add("sign", sign)
	for _, k := range keys {
		values.Add(k, ps[k])
	}
	qs = values.Encode()
	return
}

func (req *Request) Fields(fields ...string) {
	req.Params["fields"] = strings.Join(fields, ",")
}

func (req *Request) NumIids(ids ...int64) {
	var strids []string
	for _, id := range ids {
		strids = append(strids, strconv.FormatInt(id, 10))
	}
	req.Params["num_iids"] = strings.Join(strids, ",")
}

func (req *Request) Nicks(nicks ...string) {
	req.Params["nicks"] = strings.Join(nicks, ",")
}

func (req *Request) makeRequestParams() map[string]string {
	ps := map[string]string{}
	ps["v"] = req.Client.v
	ps["sign_method"] = req.Client.signMethod
	ps["app_key"] = req.Client.AppKey
	if req.Client.SessionKey != "" {
		ps["session"] = req.Client.SessionKey
	}
	ps["partner_id"] = req.Client.PartnerId
	ps["method"] = req.Name
	ps["format"] = req.Client.format
	ps["timestamp"] = time.Now().Format("2006-01-02 15:04:05")

	for k, v := range req.Params {
		ps[k] = v
	}
	return ps
}
