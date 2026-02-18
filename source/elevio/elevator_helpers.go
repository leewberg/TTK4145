package elevio

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
	"time"
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

func (e *elevator) isOrderInFloor(dir t.OrderType, floor int) bool {
	order := db.GetOrder(dir, floor)
	return order.IsActive() && order.AssignedID == e.ID && time.Now().UnixMilli()-order.AssignedTime > cfg.BiddingTime
}

func (e *elevator) viable_floor(floor int) bool {

	return e.isOrderInFloor(t.OrderType(2+e.ID), floor) || e.isOrderInFloor(MDToOrdertype(e.Direction), floor)
}

func ChooseDirection(elevData elevator, simRequests map[t.OrderType][]bool, duration int) (MotorDirection, int) {
	// check for orders in current direction of travel. if there are none, turn around
	switch elevData.Direction {
	case MD_Up:
		if RequestsAbove(elevData, simRequests) {
			return MD_Up, duration
		} else if AnyRequestsAtFloor(elevData.In_floor, simRequests) {
			return MD_Stop, duration
		} else if RequestsBelow(elevData, simRequests) {
			return MD_Down, duration + 5000
		} else {
			return MD_Stop, duration
		}
	default:
		if RequestsBelow(elevData, simRequests) {
			return MD_Down, duration
		} else if AnyRequestsAtFloor(elevData.In_floor, simRequests) {
			return MD_Stop, duration
		} else if RequestsAbove(elevData, simRequests) {
			return MD_Up, duration + 5000
		} else {
			return MD_Stop, duration
		}
	}
}

func RequestsAbove(elevData elevator, simRequests map[t.OrderType][]bool) bool {
	for floor := elevData.In_floor + 1; floor < cfg.NumFloors; floor++ {
		if AnyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func RequestsBelow(elevData elevator, simRequests map[t.OrderType][]bool) bool {
	for floor := elevData.In_floor - 1; floor >= 0; floor-- {
		if AnyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func AnyRequestsAtFloor(floor int, simRequests map[t.OrderType][]bool) bool {
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
