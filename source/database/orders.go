package database

import (
	// "fmt"
	config "heislabb/source/config"
	types "heislabb/source/types"
	"sync"
	"time"
)

var allOrdersData map[types.OrderType][]types.OrderData
var mutexOD sync.RWMutex

func InitOrderData() {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if allOrdersData == nil {
		allOrdersData = make(map[types.OrderType][]types.OrderData)
	}

	for orderType := types.HALL_UP; orderType < config.NUM_ELEVATORS+2; orderType++ {
		allOrdersData[orderType] = make([]types.OrderData, config.NUM_FLOORS)
		for floor := range config.NUM_FLOORS {
			allOrdersData[orderType][floor] = types.OrderData{Version: 0, AssignedID: -1, AssignedCost: types.INF, AssignedAtTime: 0}

		}
	}

}

func RequestOrder(orderType types.OrderType, orderFloor int) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if types.StateFromVersionNr(allOrdersData[orderType][orderFloor].Version) == types.ORDER_CLEAR {
		allOrdersData[orderType][orderFloor].Version += 1
	}
}

func ClearOrder(orderType types.OrderType, orderFloor int) {
	// fmt.Println("requesting clear @", orderType, orderFloor)
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if types.StateFromVersionNr(allOrdersData[orderType][orderFloor].Version) == types.ORDER_CONFIRMED &&
		allOrdersData[orderType][orderFloor].AssignedID == config.MY_ID {
		allOrdersData[orderType][orderFloor].Version += 1
		WorkProven()
		if isAloneOnNetwork() {
			allOrdersData[orderType][orderFloor].Version = 0
		}
	}
}

func ReadOrderData(orderType types.OrderType, orderFloor int) types.OrderData {
	mutexOD.RLock()
	defer mutexOD.RUnlock()
	return allOrdersData[orderType][orderFloor]
}

func AssignOrder(orderType types.OrderType, orderFloor int, cost int) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	isElevFunctional := GetFunctionalElevators()
	if orderType < types.CAB_FIRST { // is hall order
		if !isElevFunctional[config.MY_ID] {
			return
		}
		WorkProven()
	}

	if allOrdersData[orderType][orderFloor].GetState() == types.ORDER_REQUESTED {
		allOrdersData[orderType][orderFloor].Version += 1
		allOrdersData[orderType][orderFloor].AssignedCost = cost
		allOrdersData[orderType][orderFloor].AssignedID = config.MY_ID
		allOrdersData[orderType][orderFloor].AssignedAtTime = time.Now().UnixMilli()

	} else if allOrdersData[orderType][orderFloor].GetState() == types.ORDER_CONFIRMED {
		allOrdersData[orderType][orderFloor].AssignedCost = cost
		allOrdersData[orderType][orderFloor].AssignedID = config.MY_ID
		allOrdersData[orderType][orderFloor].AssignedAtTime = time.Now().UnixMilli()
	}
}

func validState(data types.OrderData) bool {
	if data.GetState() == types.ORDER_CONFIRMED &&
		data.AssignedID == -1 {
		return false
	}
	return true
}

func computeFullCost(orderData types.OrderData) int {
	cost := orderData.AssignedCost
	functionalElevators := GetFunctionalElevators()
	if !functionalElevators[orderData.AssignedID] {
		cost += types.INF
	}
	cost += orderData.AssignedID // use ID for tiebreaks. ensure ID < MIN_RAISE
	return cost
}

func MergeOrder(orderType types.OrderType, orderFloor int, mergeData types.OrderData) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if !validState(mergeData) {
		return
	}

	currentOrder := allOrdersData[orderType][orderFloor]

	if mergeData.Version > currentOrder.Version {

		// Stubbornness clause: you should not externally clear an order assigned to this node
		if currentOrder.GetState() == types.ORDER_CONFIRMED &&
			currentOrder.AssignedID == config.MY_ID &&
			mergeData.GetState() != types.ORDER_CONFIRMED {

			allOrdersData[orderType][orderFloor].Version = mergeData.Version + (2 - mergeData.Version%3)

		} else {

			allOrdersData[orderType][orderFloor] = mergeData

		}

	} else if mergeData.Version == currentOrder.Version &&
		mergeData.GetState() == types.ORDER_CONFIRMED {

		currentCost := computeFullCost(currentOrder)
		incomingCost := computeFullCost(mergeData)

		if currentCost > incomingCost {
			allOrdersData[orderType][orderFloor] = mergeData
		}

	}
}
