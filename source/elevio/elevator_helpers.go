package elevio

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
)

func MDToOrdertype(dir MotorDirection) t.OrderType {
	switch dir {
	case MD_Up:
		return t.HallUp
	case MD_Down:
		return t.HallDown
	}
	return 0
}

func viableFloor(dir t.OrderType, floor int, simRequests map[t.OrderType][]bool) bool {
	//potential to be merged with anyRequests. but personally, the distinction between checking all types of orders vs. checking only the direction you're going in is important enough to keep these as seperate functions. extra logic can be added in the fsm to account for the merging of these, but this would require way more logic to check for edge cases (is there a cab order but also a hall down order, but also we're going up etc etc), and would thus become unneccecary. adding exit-types is possible, but it also requires more rewriting and checking than what is present in the current system
	return simRequests[t.GetMyCab(cfg.MyID)][floor] || simRequests[dir][floor]
}

func ChooseDirection(elevData elevator, simRequests map[t.OrderType][]bool, duration int) (MotorDirection, int) {
	// check for orders in current direction of travel. if there are none, turn around
	switch elevData.Direction {
	case MD_Up:
		if requestsAbove(elevData, simRequests) {
			return MD_Up, duration
		} else if anyRequestsAtFloor(elevData.In_floor, simRequests) {
			return MD_Stop, duration
		} else if requestsBelow(elevData, simRequests) {
			return MD_Down, duration + 5000
		} else {
			return MD_Stop, duration
		}
	default:
		if requestsBelow(elevData, simRequests) {
			return MD_Down, duration
		} else if anyRequestsAtFloor(elevData.In_floor, simRequests) {
			return MD_Stop, duration
		} else if requestsAbove(elevData, simRequests) {
			return MD_Up, duration + 5000
		} else {
			return MD_Stop, duration
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
	return simRequests[t.HallDown][floor] || simRequests[t.GetMyCab(cfg.MyID)][floor] || simRequests[t.HallUp][floor]
}

func makeSimReq(ourCab t.OrderType) map[t.OrderType][]bool {
	simRequests := make(map[t.OrderType][]bool)
	for _, orderType := range []t.OrderType{t.HallUp, t.HallDown, ourCab} {
		simRequests[orderType] = make([]bool, cfg.NumFloors)
		for floor := range cfg.NumFloors {
			orderData := db.GetOrder(orderType, floor)
			if orderData.IsActive() && orderData.AssignedID == cfg.MyID {
				simRequests[orderType][floor] = true
			}
		}
	}
	return simRequests
}
