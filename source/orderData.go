package elevio

import (
	"sync"
)

type OrderState int
type OrderType int

const (
	ORDER_CLEAR OrderState = iota
	ORDER_REQUESTED
	ORDER_CONFIRMED
)

const (
	HALL_UP OrderType = iota
	HALL_DOWN
	CAB_FIRST
)

type OrderData struct {
	version_nr int // contains state info

	// only relevant in confirmed state
	assigned_to   int
	assigned_cost int
}

var allOrdersData map[OrderType][]OrderData
var mutexOD sync.RWMutex

func stateFromVersionNr(order_version_nr int) OrderState {
	if order_version_nr%3 == 0 {
		return ORDER_CLEAR
	} else if order_version_nr%3 == 1 {
		return ORDER_REQUESTED
	} else {
		return ORDER_CONFIRMED
	}
}

func computeFullCost(orderData OrderData) float64 {
	cost := float64(orderData.assigned_cost)
	functionalElevators := getFunctionalElevators()
	if !functionalElevators[orderData.assigned_to] {
		cost += INF
	}
	cost += 0.01 * float64(orderData.assigned_to) // use ID for tiebreaks
	return cost
}

func initOrderData() {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if allOrdersData == nil {
		allOrdersData = make(map[OrderType][]OrderData)
	}

	for orderType := HALL_UP; orderType < NUM_ELEVATORS+2; orderType++ {
		allOrdersData[orderType] = make([]OrderData, NUM_FLOORS)
		for floor := range NUM_FLOORS {
			allOrdersData[orderType][floor] = OrderData{version_nr: 0, assigned_to: -1, assigned_cost: INF}

		}
	}

}

func requestOrder(orderType OrderType, orderFloor int) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if stateFromVersionNr(allOrdersData[orderType][orderFloor].version_nr) == ORDER_CLEAR {
		allOrdersData[orderType][orderFloor].version_nr += 1
	}
}

func clearOrder(orderType OrderType, orderFloor int) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if stateFromVersionNr(allOrdersData[orderType][orderFloor].version_nr) == ORDER_CONFIRMED {
		allOrdersData[orderType][orderFloor].version_nr += 1
		// allOrdersData[orderType][orderFloor].assigned_cost = INF
		// allOrdersData[orderType][orderFloor].assigned_to = -1
	}
}

func readOrderData(orderType OrderType, orderFloor int) OrderData {
	mutexOD.RLock()
	defer mutexOD.RUnlock()
	return allOrdersData[orderType][orderFloor]
}

func assignOrder(orderType OrderType, orderFloor int, assignTo int, cost int) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if stateFromVersionNr(allOrdersData[orderType][orderFloor].version_nr) == ORDER_REQUESTED {
		allOrdersData[orderType][orderFloor].version_nr += 1
		allOrdersData[orderType][orderFloor].assigned_cost = cost
		allOrdersData[orderType][orderFloor].assigned_to = assignTo

	} else if stateFromVersionNr(allOrdersData[orderType][orderFloor].version_nr) == ORDER_CONFIRMED {
		isElevFunctional := getFunctionalElevators()

		if !isElevFunctional[allOrdersData[orderType][orderFloor].assigned_to] {

			allOrdersData[orderType][orderFloor].assigned_cost = cost
			allOrdersData[orderType][orderFloor].assigned_to = assignTo
		}
	}
}

func validState(data OrderData) bool {
	if stateFromVersionNr(data.version_nr) == ORDER_CONFIRMED &&
		data.assigned_to == -1 {
		return false
	}
	return true
}

func mergeOrder(orderType OrderType, orderFloor int, mergeData OrderData) {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	if !validState(mergeData) {
		return
	}

	currentOrder := allOrdersData[orderType][orderFloor]

	if mergeData.version_nr > currentOrder.version_nr {

		// Stubbornness clause: you should not externally clear an order assigned to this node
		if stateFromVersionNr(currentOrder.version_nr) == ORDER_CONFIRMED &&
			currentOrder.assigned_to == MY_ID &&
			stateFromVersionNr(mergeData.version_nr) != ORDER_CONFIRMED {

			allOrdersData[orderType][orderFloor].version_nr = mergeData.version_nr + (2 - mergeData.version_nr%3)

		} else {

			allOrdersData[orderType][orderFloor] = mergeData

		}

	} else if mergeData.version_nr == currentOrder.version_nr &&
		stateFromVersionNr(mergeData.version_nr) == ORDER_CONFIRMED {

		currentCost := computeFullCost(currentOrder)
		incomingCost := computeFullCost(mergeData)

		if currentCost > incomingCost {
			allOrdersData[orderType][orderFloor] = mergeData
		}
		// is there a chance that reassignment messes with the stubbornness?

	}
}
