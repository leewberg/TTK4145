package assignment

import (
	config "heislabb/source/config"
	db "heislabb/source/database"
	elevio "heislabb/source/elevio"
	types "heislabb/source/types"
	"time"
)

// import "fmt"

func costFunction(orderType types.OrderType, orderFloor int) int {
	// finds the cost for the elevator to do a spesific order, by simulating execution
	elevData := elevio.LocalElevator // shallow copy should be sufficient
	duration := 0
	ourCab := types.GetMyCab(config.MY_ID)

	// copy down data so we don't override the actual orders
	simRequests := make(map[types.OrderType][]bool)
	for _, orderType := range []types.OrderType{types.HALL_UP, types.HALL_DOWN, ourCab} {
		simRequests[orderType] = make([]bool, config.NUM_FLOORS)
		for floor := range config.NUM_FLOORS {
			orderData := db.ReadOrderData(orderType, floor)
			if types.StateFromVersionNr(orderData.Version) == types.ORDER_CONFIRMED &&
				orderData.AssignedID == config.MY_ID {
				simRequests[orderType][floor] = true
			}
		}
	}
	simRequests[orderType][orderFloor] = true
	if elevData.In_floor == config.NUM_ELEVATORS-1 {
		elevData.Direction = elevio.MD_Down
	} else if elevData.In_floor == 0 {
		elevData.Direction = elevio.MD_Up
	}

	// initial considerations
	switch elevData.State {
	case elevio.ELEV_BOOT:
		return types.INF
	case elevio.ELEV_DOOR_OPEN:
		duration -= config.DOOR_OPEN_TIME / 2
	default:
		elevData.Direction, _ = chooseDirection(elevData, simRequests, duration)
	}
	if elevData.Is_between_floors {
		duration += config.TRAVEL_TIME / 2
		elevData.In_floor += int(elevData.Direction)
	}

	for {
		if elevShouldStop(elevData, simRequests) {

			// clears all orders for the floor. TODO: Punish turnarounds also duing clears
			simulatedClearRequests(elevData, simRequests)
			duration += config.DOOR_OPEN_TIME
			if !simRequests[orderType][orderFloor] {
				return duration
			}
			elevData.Direction, _ = chooseDirection(elevData, simRequests, duration)
		}
		elevData.In_floor += int(elevData.Direction)
		duration += config.TRAVEL_TIME
	}
}

func simulatedClearRequests(elevData elevio.Elevator, simRequests map[types.OrderType][]bool) {
	simRequests[types.GetMyCab(config.MY_ID)][elevData.In_floor] = false
	switch elevData.Direction {
	case elevio.MD_Up:
		if simRequests[types.HALL_UP][elevData.In_floor] {
			simRequests[types.HALL_UP][elevData.In_floor] = false
		} else if !requestsAbove(elevData, simRequests) {
			simRequests[types.HALL_DOWN][elevData.In_floor] = false
		}
	case elevio.MD_Down:
		if simRequests[types.HALL_DOWN][elevData.In_floor] {
			simRequests[types.HALL_DOWN][elevData.In_floor] = false
		} else if !requestsBelow(elevData, simRequests) {
			simRequests[types.HALL_UP][elevData.In_floor] = false
		}
	default: // MD_Stop
		simRequests[types.HALL_DOWN][elevData.In_floor] = false
		simRequests[types.HALL_UP][elevData.In_floor] = false
	}
}

func requestsAbove(elevData elevio.Elevator, simRequests map[types.OrderType][]bool) bool {
	for floor := elevData.In_floor + 1; floor < config.NUM_FLOORS; floor++ {
		if anyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func requestsBelow(elevData elevio.Elevator, simRequests map[types.OrderType][]bool) bool {
	for floor := elevData.In_floor - 1; floor >= 0; floor-- {
		if anyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func anyRequests(simRequests map[types.OrderType][]bool) bool {
	for floor := range config.NUM_FLOORS {
		if anyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func anyRequestsAtFloor(floor int, simRequests map[types.OrderType][]bool) bool {
	return simRequests[types.HALL_DOWN][floor] || simRequests[types.GetMyCab(config.MY_ID)][floor] || simRequests[types.HALL_UP][floor]
}

func elevShouldStop(elevData elevio.Elevator, simRequests map[types.OrderType][]bool) (shouldStop bool) {
	// An out of bounds check failed here at index 4. so in_floor likley got to high
	shouldStop = false
	ourCab := types.GetMyCab(config.MY_ID)

	switch elevData.Direction {
	case elevio.MD_Up:
		return (simRequests[types.HALL_UP][elevData.In_floor] ||
			simRequests[ourCab][elevData.In_floor] ||
			!requestsAbove(elevData, simRequests) ||
			elevData.In_floor >= config.NUM_FLOORS-1)
	case elevio.MD_Down:
		return (simRequests[types.HALL_DOWN][elevData.In_floor] ||
			simRequests[ourCab][elevData.In_floor] ||
			!requestsBelow(elevData, simRequests) ||
			elevData.In_floor == 0)
	default: // case MD_Stop
		return true
	}

}

func chooseDirection(elevData elevio.Elevator, simRequests map[types.OrderType][]bool, duration int) (elevio.MotorDirection, int) {
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

func AssignOrders() {

	// cab orders
	for floor := range config.NUM_FLOORS {
		order := db.ReadOrderData(types.GetMyCab(config.MY_ID), floor)
		if order.GetState() == types.ORDER_REQUESTED {
			db.AssignOrder(types.GetMyCab(config.MY_ID), floor, 0)
		}
	}

	// hall orders
	for _, orderType := range []types.OrderType{types.HALL_UP, types.HALL_DOWN} {
		for floor := range config.NUM_FLOORS {

			order := db.ReadOrderData(orderType, floor)

			if order.GetState() == types.ORDER_REQUESTED ||
				(order.GetState() == types.ORDER_CONFIRMED &&
					time.Now().UnixMilli()-db.GetLastProofOfWork(order.AssignedID) > config.ELEVATOR_TIMEOUT) {

				if order.GetState() == types.ORDER_CONFIRMED {
					db.OrderFailed(order.AssignedID)
					if order.AssignedID == config.MY_ID {
						continue
					}
				}

				cost := costFunction(orderType, floor)
				db.AssignOrder(orderType, floor, cost)

			} else if order.GetState() == types.ORDER_CONFIRMED &&
				time.Now().UnixMilli()-order.AssignedAtTime < config.BIDDING_TIME {

				cost := costFunction(orderType, floor)
				// fmt.Println("Bidding with cost", cost, "on order", orderType, floor, "against", order.assigned_cost)
				if cost+config.BIDDING_MIN_RAISE < order.AssignedCost {
					// fmt.Println("Got the bid with cost", cost, "on order", orderType, floor)
					db.AssignOrder(orderType, floor, cost)
				}
			}
		}
	}
}
