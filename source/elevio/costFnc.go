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
	if elevData.inFloor == cfg.NumElevators-1 {
		elevData.direction = t.MD_Down
	} else if elevData.inFloor == 0 {
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
		elevData.inFloor += int(elevData.direction)
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
		elevData.inFloor += int(elevData.direction)
		duration += cfg.TravelTime
	}
}

func clearMatrixOrders(elevData elevator, orderMatrix map[t.OrderType][]bool) {
	orderMatrix[t.GetMyCab(cfg.MyID)][elevData.inFloor] = false
	switch elevData.direction {
	case t.MD_Up:
		if orderMatrix[t.HallUp][elevData.inFloor] {
			orderMatrix[t.HallUp][elevData.inFloor] = false
		} else if !requestsAbove(elevData, orderMatrix) {
			orderMatrix[t.HallDown][elevData.inFloor] = false
		}
	case t.MD_Down:
		if orderMatrix[t.HallDown][elevData.inFloor] {
			orderMatrix[t.HallDown][elevData.inFloor] = false
		} else if !requestsBelow(elevData, orderMatrix) {
			orderMatrix[t.HallUp][elevData.inFloor] = false
		}
	default: // MD_Stop
		orderMatrix[t.HallDown][elevData.inFloor] = false
		orderMatrix[t.HallUp][elevData.inFloor] = false
	}
}

func elevShouldStop(elevData elevator, orderMatrix map[t.OrderType][]bool) (shouldStop bool) {
	// An out of bounds check failed here at index 4. so in_floor likley got to high
	shouldStop = false
	ourCab := t.GetMyCab(cfg.MyID)

	switch elevData.direction {
	case t.MD_Up:
		return (orderMatrix[t.HallUp][elevData.inFloor] ||
			orderMatrix[ourCab][elevData.inFloor] ||
			!requestsAbove(elevData, orderMatrix) ||
			elevData.inFloor >= cfg.NumFloors-1)
	case t.MD_Down:
		return (orderMatrix[t.HallDown][elevData.inFloor] ||
			orderMatrix[ourCab][elevData.inFloor] ||
			!requestsBelow(elevData, orderMatrix) ||
			elevData.inFloor == 0)
	default: // case MD_Stop
		return true
	}
}
