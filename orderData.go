package main

import (
	"sync"
)

type OrderState int
type OrderType int

const (
	CLEAR OrderState = iota
	ACCEPTED
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
	version_nr    int // contains state info
	assigned_to   int // only relevant for confirmed state
	assigned_cost int // ditto
}

var mutexOD sync.RWMutex
var orderData map[OrderType][]OrderData

func getOrderState(order_version_nr int) OrderState {
	if order_version_nr%3 == 0 {
		return CLEAR
	} else if order_version_nr%3 == 1 {
		return ACCEPTED
	} else {
		return CONFIRMED
	}
}

func initOrderData() {
	mutexOD.Lock()
	defer mutexOD.Unlock()

	for orderType := HALL_UP; orderType < NUM_ELEVATORS+2; orderType++ {
		for floor := range NUM_FLOORS {
			orderData[orderType][floor] = OrderData{version_nr: 0, assigned_to: -1, assigned_cost: 999999}

		}
	}

}
