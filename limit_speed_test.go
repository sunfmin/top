package top

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestLimitSpeed(t *testing.T) {
	start := time.Now()
	req := newRequest("taobao.taobaoke.items.get")
	req.Client.Verbose = true

	ExtandLimitByAddAppKey("12486123", "280a8edc3a899b8a1e4cb965732d2441", 100)
	for {
		req.Param("fields", "num_iid")
		req.Param("keyword", "nike")
		req.Param("nick", "qintb8")
		req.Param("page_size", "5")

		r := []*Item{}
		_, err := req.Execute(&r)
		log.Println(err)
		if err != nil {
			topErr := err.(*Error)
			if topErr.BanSeconds() > 0 {
				panic(fmt.Sprintf("Still banned even limited speed, %+v", currentUsingAppKey))
			}
		}
		if time.Now().Sub(start) > 2*time.Minute {
			break
		}
	}
}
