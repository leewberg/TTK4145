package database

import (
	// "fmt"
	config "heislabb/source/config"
	"sync"
	"time"
)

// Module to keep track of liveliness of the elevators
var lastProofOfWork []int64
var lastFailedOrderTime []int64
var lastRecivedMsgTime int64 // for figuring out if we are isolated
var mutexLFT sync.RWMutex

func InitPeers() {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastProofOfWork = make([]int64, config.NUM_ELEVATORS)
	lastFailedOrderTime = make([]int64, config.NUM_ELEVATORS)
	lastRecivedMsgTime = 0

	for i := range config.NUM_ELEVATORS {
		lastProofOfWork[i] = 0
	}
}

func WorkProven() {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastProofOfWork[config.MY_ID] = time.Now().UnixMilli()
}

func OrderFailed(elevatorNum int) {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastFailedOrderTime[elevatorNum] = time.Now().UnixMilli()
}

func GetLastProofOfWork(elevatorNum int) int64 {
	mutexLFT.RLock()
	defer mutexLFT.RUnlock()

	return lastProofOfWork[elevatorNum]
}

func GetLastFailedTime(elevatorNum int) int64 {
	mutexLFT.RLock()
	defer mutexLFT.RUnlock()

	return lastFailedOrderTime[elevatorNum]
}

func MergePeersData(elevatorNum int, proofOfWork int64, lastFail int64) {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastProofOfWork[elevatorNum] = max(lastProofOfWork[elevatorNum], proofOfWork)
	lastFailedOrderTime[elevatorNum] = max(lastFailedOrderTime[elevatorNum], lastFail)
}

func RecivedMsg() {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastRecivedMsgTime = time.Now().UnixMilli()
}

func isAloneOnNetwork() bool {
	mutexLFT.RLock()
	defer mutexLFT.RUnlock()
	return time.Now().UnixMilli()-lastRecivedMsgTime > config.ELEVATOR_TIMEOUT
}

func GetFunctionalElevators() []bool {
	mutexLFT.RLock()
	defer mutexLFT.RUnlock()

	now := time.Now().UnixMilli()
	funcElevs := make([]bool, config.NUM_ELEVATORS)

	for elevID := range config.NUM_ELEVATORS {
		// in case all nodes but one are dead, we need NUM_ELEVATORS-1 cycles to ensure the one functional node has a chance to grab the order
		if lastFailedOrderTime[elevID] < lastProofOfWork[elevID] ||
			now-lastFailedOrderTime[elevID] > (config.NUM_ELEVATORS)*config.ELEVATOR_TIMEOUT+1000 {
			funcElevs[elevID] = true
		} else {
			funcElevs[elevID] = false
		}
	}

	return funcElevs
}
