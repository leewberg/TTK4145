package main

import "sync"

type Direction int
type ElevState int

const (
	STOP Direction = iota
	DOWN
	UP
)

const (
	BOOT ElevState = iota
	STOPBTN
	IDLE
	RUNNING
	DOOR_OPEN
)

type ElevatorData struct {
	last_floor   int
	state        ElevState
	direction    Direction
	data_version int
}

var allElevatorsData [NUM_ELEVATORS]ElevatorData
var mutexESD sync.RWMutex

func initElevatorData() {
	mutexESD.Lock()
	defer mutexESD.Unlock()
	for i := range NUM_ELEVATORS {
		allElevatorsData[i] = ElevatorData{last_floor: -1, state: BOOT, direction: STOP, data_version: 0}
	}
}

func mergeElevatorData(elevatorNum int, newData ElevatorData) {
	// Merge with external, incoming data
	mutexESD.Lock()
	defer mutexESD.Unlock()

	if elevatorNum != MY_ID {
		if allElevatorsData[MY_ID].data_version <= newData.data_version {
			allElevatorsData[MY_ID] = newData
		}

	} else {
		// Ensure elevator always has priority over its own state info
		if allElevatorsData[MY_ID].data_version < newData.data_version {
			allElevatorsData[MY_ID].data_version = newData.data_version + 1
		}
	}
}

func getElevData(elevatorNum int) ElevatorData {
	mutexESD.RLock()
	defer mutexESD.Unlock()

	return allElevatorsData[elevatorNum]
}

func setElevData(elevatorNum int, newData ElevatorData) {
	mutexESD.Lock()
	defer mutexESD.Unlock()

	allElevatorsData[elevatorNum] = newData
}
