package elevio

import (
	"sync"
	"time"
)

// Module to keep track of the last known functional times of all elevators
var lastFunctionalTimes []int64
var lastRecivedMsgTime int64 // for figuring out if we are connected to a network
var mutexLFT sync.RWMutex

func initFunctionalTimes() {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastFunctionalTimes = make([]int64, NUM_ELEVATORS)
	lastRecivedMsgTime = 0

	for i := range NUM_ELEVATORS {
		lastFunctionalTimes[i] = 0
	}
}

func declareElevatorFunctional() {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastFunctionalTimes[MY_ID] = time.Now().UnixMilli()
}

func readElevatorFunctional(elevatorNum int) int64 {
	mutexLFT.RLock()
	defer mutexLFT.RUnlock()

	return lastFunctionalTimes[elevatorNum]
}

func mergeElevFunctionalData(elevatorNum int, value int64) {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastFunctionalTimes[elevatorNum] = max(lastFunctionalTimes[elevatorNum], value)
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
		if now-lastFunctionalTimes[elevID] > ELEVATOR_TIMEOUT {
			funcElevs[elevID] = false
		} else {

			funcElevs[elevID] = true
		}
	}

	return funcElevs
}
