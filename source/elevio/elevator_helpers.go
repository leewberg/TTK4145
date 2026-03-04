package elevio

import (
	cfg "heislabb/source/config"
	t "heislabb/source/types"
)

func mdToOrdertype(dir MotorDirection) t.OrderType {
	switch dir {
	case MD_Up:
		return t.HallUp
	case MD_Down:
		return t.HallDown
	}
	return 0
}

func isOrder(dir t.OrderType, floor int, simRequests map[t.OrderType][]bool) bool {
	return simRequests[dir][floor]
}

func ChooseDirection(elevData elevator, simRequests map[t.OrderType][]bool) MotorDirection {
	// check for orders in current direction of travel. if there are none, turn around
	switch elevData.Direction {
	case MD_Up:
		if requestsAbove(elevData, simRequests) {
			return MD_Up
		} else if anyRequestsAtFloor(elevData.In_floor, simRequests) {
			return MD_Stop
		} else if requestsBelow(elevData, simRequests) {
			return MD_Down
		} else {
			return MD_Stop
		}
	default:
		if requestsBelow(elevData, simRequests) {
			return MD_Down
		} else if anyRequestsAtFloor(elevData.In_floor, simRequests) {
			return MD_Stop
		} else if requestsAbove(elevData, simRequests) {
			return MD_Up
		} else {
			return MD_Stop
		}
	}
}

func requestsAbove(elevData elevator, simRequests map[t.OrderType][]bool) bool {
	for floor := elevData.In_floor + 1; floor < cfg.NumFloors; floor++ {
		if anyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func requestsBelow(elevData elevator, simRequests map[t.OrderType][]bool) bool {
	for floor := elevData.In_floor - 1; floor >= 0; floor-- {
		if anyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func anyRequestsAtFloor(floor int, simRequests map[t.OrderType][]bool) bool {
	return isOrder(t.HallDown, floor, simRequests) || isOrder(t.GetMyCab(cfg.MyID), floor, simRequests) || isOrder(t.HallUp, floor, simRequests)
}
