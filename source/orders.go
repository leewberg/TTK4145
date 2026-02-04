package elevio

import (
	// "fmt"
	"sync"
	"time"
)

var allOrdersData map[OrderType][]OrderData
var mutexOD sync.RWMutex

func InitOrderData() {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if allOrdersData == nil {
		allOrdersData = make(map[OrderType][]OrderData)
	}

	for orderType := HALL_UP; orderType < NUM_ELEVATORS+2; orderType++ {
		allOrdersData[orderType] = make([]OrderData, NUM_FLOORS)
		for floor := range NUM_FLOORS {
			allOrdersData[orderType][floor] = OrderData{Version: 0, AssignedID: -1, AssignedCost: INF, AssignedAtTime: 0}

		}
	}

}

func RequestOrder(orderType OrderType, orderFloor int) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if StateFromVersionNr(allOrdersData[orderType][orderFloor].Version) == ORDER_CLEAR {
		allOrdersData[orderType][orderFloor].Version += 1
	}
}

func ClearOrder(orderType OrderType, orderFloor int) {
	// fmt.Println("requesting clear @", orderType, orderFloor)
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if StateFromVersionNr(allOrdersData[orderType][orderFloor].Version) == ORDER_CONFIRMED &&
		allOrdersData[orderType][orderFloor].AssignedID == MY_ID {
		allOrdersData[orderType][orderFloor].Version += 1
		workProven()
		if isAloneOnNetwork() {
			allOrdersData[orderType][orderFloor].Version = 0
		}
	}
}

func ReadOrderData(orderType OrderType, orderFloor int) OrderData {
	mutexOD.RLock()
	defer mutexOD.RUnlock()
	return allOrdersData[orderType][orderFloor]
}

func AssignOrder(orderType OrderType, orderFloor int, cost int) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	isElevFunctional := getFunctionalElevators()
	if orderType < CAB_FIRST { // is hall order
		if !isElevFunctional[MY_ID] {
			return
		}
		workProven()
	}

	if StateFromVersionNr(allOrdersData[orderType][orderFloor].Version) == ORDER_REQUESTED {
		allOrdersData[orderType][orderFloor].Version += 1
		allOrdersData[orderType][orderFloor].AssignedCost = cost
		allOrdersData[orderType][orderFloor].AssignedID = MY_ID
		allOrdersData[orderType][orderFloor].AssignedAtTime = time.Now().UnixMilli()

	} else if StateFromVersionNr(allOrdersData[orderType][orderFloor].Version) == ORDER_CONFIRMED {
		allOrdersData[orderType][orderFloor].AssignedCost = cost
		allOrdersData[orderType][orderFloor].AssignedID = MY_ID
		allOrdersData[orderType][orderFloor].AssignedAtTime = time.Now().UnixMilli()
	}
}

func validState(data OrderData) bool {
	if StateFromVersionNr(data.Version) == ORDER_CONFIRMED &&
		data.AssignedID == -1 {
		return false
	}
	return true
}

func computeFullCost(orderData OrderData) int {
	cost := orderData.AssignedCost
	functionalElevators := getFunctionalElevators()
	if !functionalElevators[orderData.AssignedID] {
		cost += INF
	}
	cost += orderData.AssignedID // use ID for tiebreaks. ensure ID < MIN_RAISE
	return cost
}

func MergeOrder(orderType OrderType, orderFloor int, mergeData OrderData) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if !validState(mergeData) {
		return
	}

	currentOrder := allOrdersData[orderType][orderFloor]

	if mergeData.Version > currentOrder.Version {

		// Stubbornness clause: you should not externally clear an order assigned to this node
		if StateFromVersionNr(currentOrder.Version) == ORDER_CONFIRMED &&
			currentOrder.AssignedID == MY_ID &&
			StateFromVersionNr(mergeData.Version) != ORDER_CONFIRMED {

			allOrdersData[orderType][orderFloor].Version = mergeData.Version + (2 - mergeData.Version%3)

		} else {

			allOrdersData[orderType][orderFloor] = mergeData

		}

	} else if mergeData.Version == currentOrder.Version &&
		StateFromVersionNr(mergeData.Version) == ORDER_CONFIRMED {

		currentCost := computeFullCost(currentOrder)
		incomingCost := computeFullCost(mergeData)

		if currentCost > incomingCost {
			allOrdersData[orderType][orderFloor] = mergeData
		}

	}
}
