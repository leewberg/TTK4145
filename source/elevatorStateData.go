package elevio

import "sync"

type Direction int
type ElevState int

const (
	DIR_UP   Direction = 1
	DIR_DOWN Direction = -1
)

const (
	STATE_BOOT ElevState = iota
	STATE_STOP
	STATE_IDLE
	STATE_RUNNING
	STATE_DOOR_OPEN
)

type ElevatorData struct {
	last_floor   int
	state        ElevState
	direction    Direction
	data_version int
}

var allElevatorsData []ElevatorData
var mutexESD sync.RWMutex

func InitElevatorData() {
	mutexESD.Lock()
	defer mutexESD.Unlock()

	allElevatorsData = make([]ElevatorData, NUM_ELEVATORS)
	for i := range NUM_ELEVATORS {
		allElevatorsData[i] = ElevatorData{last_floor: -1, state: STATE_BOOT, direction: DIR_UP, data_version: 0}
	}
}

func mergeElevatorData(elevatorNum int, newData ElevatorData) {
	// Merge with external, incoming data
	mutexESD.Lock()
	defer mutexESD.Unlock()

	if elevatorNum != MY_ID {
		if allElevatorsData[elevatorNum].data_version <= newData.data_version {
			allElevatorsData[elevatorNum] = newData
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
	defer mutexESD.RUnlock()

	return allElevatorsData[elevatorNum]
}

func setElevData(newData ElevatorData) {
	mutexESD.Lock()
	defer mutexESD.Unlock()

	allElevatorsData[MY_ID] = newData
	allElevatorsData[MY_ID].data_version += 1
}
