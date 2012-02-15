package top

import (
	"encoding/json"
)

type Client struct {
	AppKey string
	SecretKey string

	signMethod string
	format  string
	partnerId string
	v string
}

type Request struct {
	Name string
	Params map[string]string
}

func NewClient() *Client {
	c := &Client{}
	c.partnerId = "top-sdk-go-20120214"
	c.format = "json"
	c.signMethod = "md5"
	c.v = "2.0"
	return c
}

func (client *Client) NewRequest(name string) *Request {
	return &Request{Name: name, Params: map[string]string{}}
}

func (req *Request) Execute(result interface{}) {
	json.Unmarshal([]byte(`{"Name": "Wednesday", "Age": 6, "Parents": ["Gomez", "Morticia"]}`), result)
}


func (req *Request) Param(name string, value string) {
	req.Params[name] = value
}

