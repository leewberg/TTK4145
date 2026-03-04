package elevio

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
)

func CostFunction(orderType t.OrderType, orderFloor int) int {
	// finds the cost for the elevator to do a spesific order, by simulating execution
	elevData := LocalElevator // shallow copy should be sufficient
	duration := 0
	ourCab := t.GetMyCab(cfg.MyID)

	orderMatrix := db.GetOrderMatrix(ourCab)
	orderMatrix[orderType][orderFloor] = true

	//bounds check
	if elevData.lastFloor == cfg.NumFloors-1 {
		elevData.direction = t.MD_Down
	} else if elevData.lastFloor == 0 {
		elevData.direction = t.MD_Up
	}

	// initial considerations
	switch elevData.state {
	case t.ELEV_BOOT:
		return t.INF
	case t.ELEV_DOOR_OPEN:
		duration -= cfg.DoorOpenTime / 2
	default:
		elevData.direction = chooseDirection(elevData, orderMatrix)
	}
	if elevData.isBetweenFloors {
		duration += cfg.TravelTime / 2
		elevData.lastFloor += int(elevData.direction)
	}

	for {
		if elevShouldStop(elevData, orderMatrix) {

			// clears all orders for the floor
			clearMatrixOrders(elevData, orderMatrix)
			duration += cfg.DoorOpenTime
			if !orderMatrix[orderType][orderFloor] {
				return duration
			}
			elevData.direction = chooseDirection(elevData, orderMatrix)
		}
		elevData.lastFloor += int(elevData.direction)
		duration += cfg.TravelTime
	}
}

func clearMatrixOrders(elevData elevator, orderMatrix map[t.OrderType][]bool) {
	orderMatrix[t.GetMyCab(cfg.MyID)][elevData.lastFloor] = false
	switch elevData.direction {
	case t.MD_Up:
		if orderMatrix[t.HallUp][elevData.lastFloor] {
			orderMatrix[t.HallUp][elevData.lastFloor] = false
		} else if !requestsAbove(elevData, orderMatrix) {
			orderMatrix[t.HallDown][elevData.lastFloor] = false
		}
	case t.MD_Down:
		if orderMatrix[t.HallDown][elevData.lastFloor] {
			orderMatrix[t.HallDown][elevData.lastFloor] = false
		} else if !requestsBelow(elevData, orderMatrix) {
			orderMatrix[t.HallUp][elevData.lastFloor] = false
		}
	default: // MD_Stop
		orderMatrix[t.HallDown][elevData.lastFloor] = false
		orderMatrix[t.HallUp][elevData.lastFloor] = false
	}
}
