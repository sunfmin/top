package top

import (
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
	return s[i].CurrentMinuteCalledCount > s[j].CurrentMinuteCalledCount
}

var appKeyCallCountList KeyCallCounts

var currentUsingAppKey *appKeyCallCount

func (kcc *appKeyCallCount) reachingLimit() bool {
	return kcc.CurrentMinuteCalledCount >= kcc.CallTimesPerMinute-5
}

func (kcc *appKeyCallCount) updateClientKey(client *Client) {
	client.AppKey = kcc.AppKey
	client.SecretKey = kcc.SecretKey
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
	return time.Duration(int64(1*time.Minute) - int64(time.Now().Sub(kcc.CurrentMinuteStart)))
}

func ExtandLimitByAddAppKey(appkey string, secretkey string, calltimesPerMinute int) {
	appKeyCallCountList = append(appKeyCallCountList, &appKeyCallCount{
		AppKey:             appkey,
		SecretKey:          secretkey,
		CallTimesPerMinute: calltimesPerMinute,
	})
}

func countOrSwitchOrWait(client *Client) {
	if len(appKeyCallCountList) == 0 {
		return
	}

	if currentUsingAppKey == nil {
		currentUsingAppKey = appKeyCallCountList[0]
		currentUsingAppKey.CurrentMinuteStart = time.Now()
		currentUsingAppKey.updateClientKey(client)
	}
	currentUsingAppKey.mutex.Lock()
	defer currentUsingAppKey.mutex.Unlock()

	if currentUsingAppKey.reachingLimit() {
		sort.Sort(appKeyCallCountList)
		currentUsingAppKey = appKeyCallCountList[0]
		currentUsingAppKey.updateClientKey(client)
	}

	if currentUsingAppKey.reachingLimit() {
		if client.Verbose {
			log.Printf("Reaching invoke count %d, will wait for %d seconds for not to be banned.",
				currentUsingAppKey.CurrentMinuteCalledCount,
				currentUsingAppKey.waitNanoseconds()/1e9)
		}
		time.Sleep(currentUsingAppKey.waitNanoseconds())
	}

	currentUsingAppKey.count()
}
