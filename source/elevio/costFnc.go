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
	simRequests := make(map[t.OrderType][]bool)
	for _, orderType := range []t.OrderType{t.HallUp, t.HallDown, ourCab} {
		simRequests[orderType] = make([]bool, cfg.NumFloors)
		for floor := range cfg.NumFloors {
			orderData := db.GetOrder(orderType, floor)
			if orderData.IsActive() &&
				orderData.AssignedID == cfg.MyID {
				simRequests[orderType][floor] = true
			}
		}
	}
	simRequests[orderType][orderFloor] = true
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
		elevData.Direction, _ = ChooseDirection(elevData, simRequests, duration)
	}
	if elevData.Is_between_floors {
		duration += cfg.TravelTime / 2
		elevData.In_floor += int(elevData.Direction)
	}

	for {
		if elevShouldStop(elevData, simRequests) {

			// clears all orders for the floor
			simulatedClearRequests(elevData, simRequests)
			duration += cfg.DoorOpenTime
			if !simRequests[orderType][orderFloor] {
				return duration
			}
			elevData.Direction, _ = ChooseDirection(elevData, simRequests, duration)
		}
		elevData.In_floor += int(elevData.Direction)
		duration += cfg.TravelTime
	}
}

func simulatedClearRequests(elevData elevator, simRequests map[t.OrderType][]bool) {
	simRequests[t.GetMyCab(cfg.MyID)][elevData.In_floor] = false
	switch elevData.Direction {
	case MD_Up:
		if simRequests[t.HallUp][elevData.In_floor] {
			simRequests[t.HallUp][elevData.In_floor] = false
		} else if !requestsAbove(elevData, simRequests) {
			simRequests[t.HallDown][elevData.In_floor] = false
		}
	case MD_Down:
		if simRequests[t.HallDown][elevData.In_floor] {
			simRequests[t.HallDown][elevData.In_floor] = false
		} else if !requestsBelow(elevData, simRequests) {
			simRequests[t.HallUp][elevData.In_floor] = false
		}
	default: // MD_Stop
		simRequests[t.HallDown][elevData.In_floor] = false
		simRequests[t.HallUp][elevData.In_floor] = false
	}
}

func elevShouldStop(elevData elevator, simRequests map[t.OrderType][]bool) (shouldStop bool) {
	// An out of bounds check failed here at index 4. so in_floor likley got to high
	shouldStop = false
	ourCab := t.GetMyCab(cfg.MyID)

	switch elevData.Direction {
	case MD_Up:
		return (simRequests[t.HallUp][elevData.In_floor] ||
			simRequests[ourCab][elevData.In_floor] ||
			!requestsAbove(elevData, simRequests) ||
			elevData.In_floor >= cfg.NumFloors-1)
	case MD_Down:
		return (simRequests[t.HallDown][elevData.In_floor] ||
			simRequests[ourCab][elevData.In_floor] ||
			!requestsBelow(elevData, simRequests) ||
			elevData.In_floor == 0)
	default: // case MD_Stop
		return true
	}

}
