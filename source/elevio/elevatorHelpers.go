package elevio

import (
	cfg "heislabb/source/config"
	t "heislabb/source/types"
)

func mdToOrdertype(dir t.MotorDirection) t.OrderType {
	switch dir {
	case t.MD_Up:
		return t.HallUp
	case t.MD_Down:
		return t.HallDown
	}
	return 0
}

func isOrder(dir t.OrderType, floor int, orderMatrix map[t.OrderType][]bool) bool {
	return orderMatrix[dir][floor]
}

func chooseDirection(elevData elevator, orderMatrix map[t.OrderType][]bool) t.MotorDirection {
	// check for orders in current direction of travel. if there are none, turn around
	switch elevData.direction {
	case t.MD_Up:
		if requestsAbove(elevData, orderMatrix) {
			return t.MD_Up
		} else if anyRequestsAtFloor(elevData.lastFloor, orderMatrix) {
			return t.MD_Stop
		} else if requestsBelow(elevData, orderMatrix) {
			return t.MD_Down
		} else {
			return t.MD_Stop
		}
	default:
		if requestsBelow(elevData, orderMatrix) {
			return t.MD_Down
		} else if anyRequestsAtFloor(elevData.lastFloor, orderMatrix) {
			return t.MD_Stop
		} else if requestsAbove(elevData, orderMatrix) {
			return t.MD_Up
		} else {
			return t.MD_Stop
		}
	}
}

func elevShouldStop(elev elevator, orderMatrix map[t.OrderType][]bool) bool {
	ourCab := t.GetMyCab(cfg.MyID)

	switch elev.direction {
	case t.MD_Up:
		return (orderMatrix[t.HallUp][elev.lastFloor] ||
			orderMatrix[ourCab][elev.lastFloor] ||
			!requestsAbove(elev, orderMatrix) ||
			elev.lastFloor >= cfg.NumFloors-1)
	case t.MD_Down:
		return (orderMatrix[t.HallDown][elev.lastFloor] ||
			orderMatrix[ourCab][elev.lastFloor] ||
			!requestsBelow(elev, orderMatrix) ||
			elev.lastFloor == 0)
	default: // case MD_Stop
		return true
	}
}

func requestsAbove(elevData elevator, orderMatrix map[t.OrderType][]bool) bool {
	for floor := elevData.lastFloor + 1; floor < cfg.NumFloors; floor++ {
		if anyRequestsAtFloor(floor, orderMatrix) {
			return true
		}
	}
	return false
}

func requestsBelow(elevData elevator, orderMatrix map[t.OrderType][]bool) bool {
	for floor := elevData.lastFloor - 1; floor >= 0; floor-- {
		if anyRequestsAtFloor(floor, orderMatrix) {
			return true
		}
	}
	return false
}

func anyRequestsAtFloor(floor int, orderMatrix map[t.OrderType][]bool) bool {
	return isOrder(t.HallDown, floor, orderMatrix) || isOrder(t.GetMyCab(cfg.MyID), floor, orderMatrix) || isOrder(t.HallUp, floor, orderMatrix)
}
