package main

import "time"
import "sync"

// Module to keep track of the last known functional times of all elevators
var lastFunctionalTimes [NUM_ELEVATORS]int64
var mutexLFT sync.RWMutex

func initFunctionalTimes() {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

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
	defer mutexLFT.Unlock()

	return lastFunctionalTimes[elevatorNum]
}

func mergeElevFunctionalData(elevatorNum int, value int64) {
	mutexLFT.Lock()
	defer mutexLFT.Unlock()

	lastFunctionalTimes[elevatorNum] = max(lastFunctionalTimes[elevatorNum], value)
}
