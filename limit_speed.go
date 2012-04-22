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
	distance := int64(time.Now().Sub(kcc.CurrentMinuteStart))
	waitMinutes := distance / int64(time.Minute)
	return time.Duration((waitMinutes+1)*int64(time.Minute) - distance)
}

func ExtandLimitByAddAppKey(appkey string, secretkey string, calltimesPerMinute int) {
	appKeyCallCountList = append(appKeyCallCountList, &appKeyCallCount{
		AppKey:             appkey,
		SecretKey:          secretkey,
		CallTimesPerMinute: calltimesPerMinute,
	})
}

func detectBanned(err error) (rerr error) {
	rerr = err
	if len(appKeyCallCountList) <= 1 {
		return
	}
	if currentUsingAppKey == nil {
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

	currentUsingAppKey.CurrentMinuteStart = time.Now().Add(time.Duration(topError.BanSeconds()) * time.Second)
	currentUsingAppKey.CurrentMinuteCalledCount = currentUsingAppKey.CallTimesPerMinute
	rerr = errors.New("ban settled.")
	return
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
		if Verbose {
			log.Printf("Reaching invoke count %d, will wait for %d seconds for not to be banned.",
				currentUsingAppKey.CurrentMinuteCalledCount,
				currentUsingAppKey.waitNanoseconds()/1e9)
		}
		time.Sleep(currentUsingAppKey.waitNanoseconds())
	}

	currentUsingAppKey.count()
}
