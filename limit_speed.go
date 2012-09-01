package top

import (
	"errors"
	"log"
	"sort"
	"sync"
	"time"
)

type appKeyCallCount struct {
	AppKey                   string
	SecretKey                string
	CallTimesPerMinute       int
	CurrentMinuteCalledCount int
	CurrentMinuteStart       time.Time
	mutex                    sync.Mutex
}

type KeyCallCounts []*appKeyCallCount

func (s KeyCallCounts) Len() int      { return len(s) }
func (s KeyCallCounts) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s KeyCallCounts) Less(i, j int) bool {
	return s[i].CurrentMinuteCalledCount < s[j].CurrentMinuteCalledCount
}

func (kcc *appKeyCallCount) reachingLimit() bool {
	return kcc.CurrentMinuteCalledCount >= kcc.CallTimesPerMinute-5
}

func (client *Client) updateClientKey() {
	client.AppKey = client.currentUsingAppKey.AppKey
	client.SecretKey = client.currentUsingAppKey.SecretKey
}

func (kcc *appKeyCallCount) count() {
	if time.Now().Sub(kcc.CurrentMinuteStart) > 1*time.Minute {
		kcc.CurrentMinuteCalledCount = 0
		kcc.CurrentMinuteStart = time.Now()
	}
	kcc.CurrentMinuteCalledCount++
	return
}

func (kcc *appKeyCallCount) waitNanoseconds() time.Duration {
	distance := int64(time.Now().Sub(kcc.CurrentMinuteStart))
	waitMinutes := distance / int64(time.Minute)
	return time.Duration((waitMinutes+1)*int64(time.Minute) - distance)
}

func (client *Client) ExtandLimitByAddAppKey(appkey string, secretkey string, calltimesPerMinute int) {
	client.appKeyCallCountList = append(client.appKeyCallCountList, &appKeyCallCount{
		AppKey:             appkey,
		SecretKey:          secretkey,
		CallTimesPerMinute: calltimesPerMinute,
	})
}

func (client *Client) switchedKeyIfBanned(err error) (rerr error, switched bool) {
	rerr = err
	if len(client.appKeyCallCountList) <= 1 {
		return
	}
	if client.currentUsingAppKey == nil {
		return
	}
	topError, ok := err.(*Error)
	if !ok {
		return
	}
	if topError.BanSeconds() == 0 {
		return
	}
	if Verbose {
		log.Printf("settling ban for %+v", topError)
	}

	client.currentUsingAppKey.CurrentMinuteStart = time.Now().Add(time.Duration(topError.BanSeconds()) * time.Second)
	client.currentUsingAppKey.CurrentMinuteCalledCount = client.currentUsingAppKey.CallTimesPerMinute
	rerr = errors.New("ban settled.")
	switched = client.countOrSwitchOrWait()
	return
}

func (client *Client) countOrSwitchOrWait() (switched bool) {
	if len(client.appKeyCallCountList) == 0 {
		return
	}

	if client.currentUsingAppKey == nil {
		client.currentUsingAppKey = client.appKeyCallCountList[0]
		client.currentUsingAppKey.CurrentMinuteStart = time.Now()
		client.updateClientKey()
	}
	client.currentUsingAppKey.mutex.Lock()
	defer client.currentUsingAppKey.mutex.Unlock()

	if client.currentUsingAppKey.reachingLimit() {
		fromKey := client.currentUsingAppKey.AppKey
		sort.Sort(client.appKeyCallCountList)
		client.currentUsingAppKey = client.appKeyCallCountList[0]
		client.updateClientKey()
		if Verbose {
			log.Printf("AppKey switched from %s to %s.", fromKey, client.currentUsingAppKey.AppKey)
		}
	}

	if client.currentUsingAppKey.reachingLimit() {
		if Verbose {
			log.Printf("Reaching invoke count %d, will wait for %d seconds for not to be banned.",
				client.currentUsingAppKey.CurrentMinuteCalledCount,
				client.currentUsingAppKey.waitNanoseconds()/1e9)
		}
		time.Sleep(client.currentUsingAppKey.waitNanoseconds())
	}
	switched = true
	client.currentUsingAppKey.count()
	return
}
