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
	if elevData.In_floor == cfg.NumElevators-1 {
		elevData.Direction = MD_Down
	} else if elevData.In_floor == 0 {
		elevData.Direction = MD_Up
	}

	// initial considerations
	switch elevData.State {
	case t.ELEV_BOOT:
		return t.INF
	case t.ELEV_DOOR_OPEN:
		duration -= cfg.DoorOpenTime / 2
	default:
		elevData.Direction = ChooseDirection(elevData, orderMatrix)
	}
	if elevData.Is_between_floors {
		duration += cfg.TravelTime / 2
		elevData.In_floor += int(elevData.Direction)
	}

	for {
		if elevShouldStop(elevData, orderMatrix) {

			// clears all orders for the floor
			simulatedClearRequests(elevData, orderMatrix)
			duration += cfg.DoorOpenTime
			if !orderMatrix[orderType][orderFloor] {
				return duration
			}
			elevData.Direction = ChooseDirection(elevData, orderMatrix)
		}
		elevData.In_floor += int(elevData.Direction)
		duration += cfg.TravelTime
	}
}

func simulatedClearRequests(elevData elevator, orderMatrix map[t.OrderType][]bool) {
	orderMatrix[t.GetMyCab(cfg.MyID)][elevData.In_floor] = false
	switch elevData.Direction {
	case MD_Up:
		if orderMatrix[t.HallUp][elevData.In_floor] {
			orderMatrix[t.HallUp][elevData.In_floor] = false
		} else if !requestsAbove(elevData, orderMatrix) {
			orderMatrix[t.HallDown][elevData.In_floor] = false
		}
	case MD_Down:
		if orderMatrix[t.HallDown][elevData.In_floor] {
			orderMatrix[t.HallDown][elevData.In_floor] = false
		} else if !requestsBelow(elevData, orderMatrix) {
			orderMatrix[t.HallUp][elevData.In_floor] = false
		}
	default: // MD_Stop
		orderMatrix[t.HallDown][elevData.In_floor] = false
		orderMatrix[t.HallUp][elevData.In_floor] = false
	}
}

func elevShouldStop(elevData elevator, orderMatrix map[t.OrderType][]bool) (shouldStop bool) {
	// An out of bounds check failed here at index 4. so in_floor likley got to high
	shouldStop = false
	ourCab := t.GetMyCab(cfg.MyID)

	switch elevData.Direction {
	case MD_Up:
		return (orderMatrix[t.HallUp][elevData.In_floor] ||
			orderMatrix[ourCab][elevData.In_floor] ||
			!requestsAbove(elevData, orderMatrix) ||
			elevData.In_floor >= cfg.NumFloors-1)
	case MD_Down:
		return (orderMatrix[t.HallDown][elevData.In_floor] ||
			orderMatrix[ourCab][elevData.In_floor] ||
			!requestsBelow(elevData, orderMatrix) ||
			elevData.In_floor == 0)
	default: // case MD_Stop
		return true
	}

}
