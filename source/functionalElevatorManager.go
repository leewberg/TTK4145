package elevio

import (
	// "fmt"
	"sync"
	"time"
)

// Module to keep track of the last known functional times of all elevators
var lastProofOfWork []int64
var lastFailedOrderTime []int64
var lastRecivedMsgTime int64 // for figuring out if we are connected to a network. maybe isolate this
var mutexLFT sync.RWMutex

func InitFunctionalTimes() {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastProofOfWork = make([]int64, NUM_ELEVATORS)
	lastFailedOrderTime = make([]int64, NUM_ELEVATORS)
	lastRecivedMsgTime = 0

	for i := range NUM_ELEVATORS {
		lastProofOfWork[i] = 0
	}
}

func workProven() {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastProofOfWork[MY_ID] = time.Now().UnixMilli()
}

func orderFailed(elevatorNum int) {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastFailedOrderTime[elevatorNum] = time.Now().UnixMilli()
}

func getLastProofOfWork(elevatorNum int) int64 {
	mutexLFT.RLock()
	defer mutexLFT.RUnlock()

	return lastProofOfWork[elevatorNum]
}

func getLastFailedTime(elevatorNum int) int64 {
	mutexLFT.RLock()
	defer mutexLFT.RUnlock()

	return lastFailedOrderTime[elevatorNum]
}

func mergeElevFunctionalData(elevatorNum int, proofOfWork int64, lastFail int64) {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastProofOfWork[elevatorNum] = max(lastProofOfWork[elevatorNum], proofOfWork)
	lastFailedOrderTime[elevatorNum] = max(lastFailedOrderTime[elevatorNum], lastFail)
}

func recivedMsg() {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastRecivedMsgTime = time.Now().UnixMilli()
}

func isAloneOnNetwork() bool {
	// includes stuck nodes
	mutexLFT.RLock()
	defer mutexLFT.RUnlock()
	return time.Now().UnixMilli()-lastRecivedMsgTime > ELEVATOR_TIMEOUT
}

func getFunctionalElevators() []bool {
	mutexLFT.RLock()
	defer mutexLFT.RUnlock()

	now := time.Now().UnixMilli()
	funcElevs := make([]bool, NUM_ELEVATORS)

	for elevID := range NUM_ELEVATORS {
		// in case all nodes but one are dead, we need NUM_ELEVATORS-1 cycles to ensure the one functional node has a chance to grab the order
		if lastFailedOrderTime[elevID] < lastProofOfWork[elevID] ||
			now-lastFailedOrderTime[elevID] > (NUM_ELEVATORS-1)*ELEVATOR_TIMEOUT+1000 {
			funcElevs[elevID] = true
		} else {
			funcElevs[elevID] = false
		}
	}

	return funcElevs
}
