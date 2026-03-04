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

	// copy down data so we don't override the actual orders
	orderMatrix := db.GetOrderMatrix(ourCab)

	orderMatrix[orderType][orderFloor] = true
	if elevData.in_floor == cfg.NumElevators-1 {
		elevData.direction = MD_Down
	} else if elevData.in_floor == 0 {
		elevData.direction = MD_Up
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
	if elevData.is_between_floors {
		duration += cfg.TravelTime / 2
		elevData.in_floor += int(elevData.direction)
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
		elevData.in_floor += int(elevData.direction)
		duration += cfg.TravelTime
	}
}

func clearMatrixOrders(elevData elevator, orderMatrix map[t.OrderType][]bool) {
	orderMatrix[t.GetMyCab(cfg.MyID)][elevData.in_floor] = false
	switch elevData.direction {
	case MD_Up:
		if orderMatrix[t.HallUp][elevData.in_floor] {
			orderMatrix[t.HallUp][elevData.in_floor] = false
		} else if !requestsAbove(elevData, orderMatrix) {
			orderMatrix[t.HallDown][elevData.in_floor] = false
		}
	case MD_Down:
		if orderMatrix[t.HallDown][elevData.in_floor] {
			orderMatrix[t.HallDown][elevData.in_floor] = false
		} else if !requestsBelow(elevData, orderMatrix) {
			orderMatrix[t.HallUp][elevData.in_floor] = false
		}
	default: // MD_Stop
		orderMatrix[t.HallDown][elevData.in_floor] = false
		orderMatrix[t.HallUp][elevData.in_floor] = false
	}
}

func elevShouldStop(elevData elevator, orderMatrix map[t.OrderType][]bool) (shouldStop bool) {
	// An out of bounds check failed here at index 4. so in_floor likley got to high
	shouldStop = false
	ourCab := t.GetMyCab(cfg.MyID)

	switch elevData.direction {
	case MD_Up:
		return (orderMatrix[t.HallUp][elevData.in_floor] ||
			orderMatrix[ourCab][elevData.in_floor] ||
			!requestsAbove(elevData, orderMatrix) ||
			elevData.in_floor >= cfg.NumFloors-1)
	case MD_Down:
		return (orderMatrix[t.HallDown][elevData.in_floor] ||
			orderMatrix[ourCab][elevData.in_floor] ||
			!requestsBelow(elevData, orderMatrix) ||
			elevData.in_floor == 0)
	default: // case MD_Stop
		return true
	}
}
