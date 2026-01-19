package main

import (
	"sync"
)

type OrderState int
type OrderType int

const (
	CLEAR OrderState = iota
	REQUESTED
	CONFIRMED
)

const (
	HALL_UP OrderType = iota
	HALL_DOWN
	CAB_1
	CAB_2
	CAB_3 // TODO: find a way to automate this
)

type OrderData struct {
	version_nr int // contains state info

	// only relevant for confirmed state
	assigned_to   int
	assigned_cost int
}

var allOrdersData map[OrderType][]OrderData
var mutexOD sync.RWMutex

func orderVersion2State(order_version_nr int) OrderState {
	if order_version_nr%3 == 0 {
		return CLEAR
	} else if order_version_nr%3 == 1 {
		return REQUESTED
	} else {
		return CONFIRMED
	}
}

func initOrderData() {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	for orderType := HALL_UP; orderType < NUM_ELEVATORS+2; orderType++ {
		for floor := range NUM_FLOORS {
			allOrdersData[orderType][floor] = OrderData{version_nr: 0, assigned_to: -1, assigned_cost: INF}

		}
	}

}

func requestOrder(orderType OrderType, orderFloor OrderType) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if orderVersion2State(allOrdersData[orderType][orderFloor].version_nr) == CLEAR {
		allOrdersData[orderType][orderFloor].version_nr += 1
	}
}

func clearOrder(orderType OrderType, orderFloor OrderType) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if orderVersion2State(allOrdersData[orderType][orderFloor].version_nr) == CONFIRMED {
		allOrdersData[orderType][orderFloor].version_nr += 1
	}
}

func readOrderData(orderType OrderType, orderFloor OrderType) OrderData {
	mutexOD.RLock()
	defer mutexOD.Unlock()
	return allOrdersData[orderType][orderFloor]
}

func mergeOrder(orderType OrderType, orderFloor OrderType, mergeData OrderData) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	currentOrder := allOrdersData[orderType][orderFloor] // readability dummy

	if mergeData.version_nr > currentOrder.version_nr {

		// Stubbornnes clause: you should not externally clear an order assigned to this node
		if orderVersion2State(currentOrder.version_nr) == CONFIRMED &&
			currentOrder.assigned_to == MY_ID &&
			orderVersion2State(mergeData.version_nr) != CONFIRMED {

			allOrdersData[orderType][orderFloor].version_nr = mergeData.version_nr + (2 - mergeData.version_nr%3)

		} else {
			allOrdersData[orderType][orderFloor] = mergeData
		}

	} else if mergeData.version_nr == currentOrder.version_nr &&
		orderVersion2State(mergeData.version_nr) == CONFIRMED {

		// Lowest cost gets the order
		if currentOrder.assigned_cost > mergeData.assigned_cost {

			allOrdersData[orderType][orderFloor] = mergeData

		} else if currentOrder.assigned_cost == mergeData.assigned_cost {

			// Elevator id is the tiebreaker
			if currentOrder.assigned_to > mergeData.assigned_to {
				allOrdersData[orderType][orderFloor] = mergeData
			}
		}
	}
}
