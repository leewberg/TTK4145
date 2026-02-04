package assignment

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	elevio "heislabb/source/elevio"
	t "heislabb/source/types"
)

func costFunction(orderType t.OrderType, orderFloor int) int {
	// finds the cost for the elevator to do a spesific order, by simulating execution
	elevData := elevio.LocalElevator // shallow copy should be sufficient
	duration := 0
	ourCab := t.GetMyCab(cfg.MyID)

	// copy down data so we don't override the actual orders
	simRequests := make(map[t.OrderType][]bool)
	for _, orderType := range []t.OrderType{t.HallUp, t.HallDown, ourCab} {
		simRequests[orderType] = make([]bool, cfg.NumFloors)
		for floor := range cfg.NumFloors {
			orderData := db.GetOrder(orderType, floor)
			if t.StateFromVersionNr(orderData.Version) == t.Confirmed &&
				orderData.AssignedID == cfg.MyID {
				simRequests[orderType][floor] = true
			}
		}
	}
	simRequests[orderType][orderFloor] = true
	if elevData.In_floor == cfg.NumElevators-1 {
		elevData.Direction = elevio.MD_Down
	} else if elevData.In_floor == 0 {
		elevData.Direction = elevio.MD_Up
	}

	// initial considerations
	switch elevData.State {
	case elevio.ELEV_BOOT:
		return t.INF
	case elevio.ELEV_DOOR_OPEN:
		duration -= cfg.DoorOpenTime / 2
	default:
		elevData.Direction, _ = chooseDirection(elevData, simRequests, duration)
	}
	if elevData.Is_between_floors {
		duration += cfg.TravelTime / 2
		elevData.In_floor += int(elevData.Direction)
	}

	for {
		if elevShouldStop(elevData, simRequests) {

			// clears all orders for the floor. TODO: Punish turnarounds also duing clears
			simulatedClearRequests(elevData, simRequests)
			duration += cfg.DoorOpenTime
			if !simRequests[orderType][orderFloor] {
				return duration
			}
			elevData.Direction, _ = chooseDirection(elevData, simRequests, duration)
		}
		elevData.In_floor += int(elevData.Direction)
		duration += cfg.TravelTime
	}
}

func simulatedClearRequests(elevData elevio.Elevator, simRequests map[t.OrderType][]bool) {
	simRequests[t.GetMyCab(cfg.MyID)][elevData.In_floor] = false
	switch elevData.Direction {
	case elevio.MD_Up:
		if simRequests[t.HallUp][elevData.In_floor] {
			simRequests[t.HallUp][elevData.In_floor] = false
		} else if !requestsAbove(elevData, simRequests) {
			simRequests[t.HallDown][elevData.In_floor] = false
		}
	case elevio.MD_Down:
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

func requestsAbove(elevData elevio.Elevator, simRequests map[t.OrderType][]bool) bool {
	for floor := elevData.In_floor + 1; floor < cfg.NumFloors; floor++ {
		if anyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func requestsBelow(elevData elevio.Elevator, simRequests map[t.OrderType][]bool) bool {
	for floor := elevData.In_floor - 1; floor >= 0; floor-- {
		if anyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func anyRequests(simRequests map[t.OrderType][]bool) bool {
	for floor := range cfg.NumFloors {
		if anyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func anyRequestsAtFloor(floor int, simRequests map[t.OrderType][]bool) bool {
	return simRequests[t.HallDown][floor] || simRequests[t.GetMyCab(cfg.MyID)][floor] || simRequests[t.HallUp][floor]
}

func elevShouldStop(elevData elevio.Elevator, simRequests map[t.OrderType][]bool) (shouldStop bool) {
	// An out of bounds check failed here at index 4. so in_floor likley got to high
	shouldStop = false
	ourCab := t.GetMyCab(cfg.MyID)

	switch elevData.Direction {
	case elevio.MD_Up:
		return (simRequests[t.HallUp][elevData.In_floor] ||
			simRequests[ourCab][elevData.In_floor] ||
			!requestsAbove(elevData, simRequests) ||
			elevData.In_floor >= cfg.NumFloors-1)
	case elevio.MD_Down:
		return (simRequests[t.HallDown][elevData.In_floor] ||
			simRequests[ourCab][elevData.In_floor] ||
			!requestsBelow(elevData, simRequests) ||
			elevData.In_floor == 0)
	default: // case MD_Stop
		return true
	}

}

func chooseDirection(elevData elevio.Elevator, simRequests map[t.OrderType][]bool, duration int) (elevio.MotorDirection, int) {
	// check for orders in current direction of travel. if there are none, turn around
	switch elevData.Direction {
	case elevio.MD_Up:
		if requestsAbove(elevData, simRequests) {
			return elevio.MD_Up, duration
		} else if anyRequestsAtFloor(elevData.In_floor, simRequests) {
			return elevio.MD_Stop, duration
		} else if requestsBelow(elevData, simRequests) {
			return elevio.MD_Down, duration + 5000
		} else {
			return elevio.MD_Stop, duration
		}
	default:
		if requestsBelow(elevData, simRequests) {
			return elevio.MD_Down, duration
		} else if anyRequestsAtFloor(elevData.In_floor, simRequests) {
			return elevio.MD_Stop, duration
		} else if requestsAbove(elevData, simRequests) {
			return elevio.MD_Up, duration + 5000
		} else {
			return elevio.MD_Stop, duration
		}
	}
}
