package elevio

import (
	"sync"
	"time"
)

// Module to keep track of the last known functional times of all elevators
var lastFunctionalTimes []int64
var mutexLFT sync.RWMutex

func initFunctionalTimes() {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastFunctionalTimes = make([]int64, NUM_ELEVATORS)

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
