package elevio

import "time"

// import "fmt"

func forceInBounds(x int, lb int, ub int) int {
	if x < lb {
		return lb
	} else if x > ub {
		return ub
	} else {
		return x
	}
}

func costFunction(orderType OrderType, orderFloor int) int {
	// finds the cost for the elevator to do a spesific order, by simulating execution
	elevData := LocalElevator // shallow copy should be sufficient
	duration := 0
	ourCab := OrderType(2 + MY_ID)

	// copy down data so we don't override the actual orders
	simRequests := make(map[OrderType][]bool)
	for _, orderType := range []OrderType{HALL_UP, HALL_DOWN, ourCab} {
		simRequests[orderType] = make([]bool, NUM_FLOORS)
		for floor := range NUM_FLOORS {
			orderData := ReadOrderData(orderType, floor)
			if stateFromVersionNr(orderData.version_nr) == ORDER_CONFIRMED &&
				orderData.assigned_to == MY_ID {
				simRequests[orderType][floor] = true
			}
		}
	}
	simRequests[orderType][orderFloor] = true
	if elevData.in_floor == NUM_ELEVATORS-1 {
		elevData.direction = MD_Down
	} else if elevData.in_floor == 0 {
		elevData.direction = MD_Up
	}

	// initial considerations
	switch elevData.state {
	case ELEV_BOOT:
		return INF
	case ELEV_DOOR_OPEN:
		duration -= DOOR_OPEN_TIME / 2
	default:
		elevData.direction, _ = chooseDirection(elevData, simRequests, ourCab, duration)
	}
	if elevData.is_between_floors {
		duration += TRAVEL_TIME / 2
		elevData.in_floor += int(elevData.direction)
	}

	for {
		if elevShouldStop(elevData, simRequests, ourCab) {

			// clears all orders for the floor. TODO: Punish turnarounds also duing clears
			simulatedClearRequests(elevData, simRequests, ourCab)
			duration += DOOR_OPEN_TIME
			if !simRequests[orderType][orderFloor] {
				return duration
			}
			elevData.direction, _ = chooseDirection(elevData, simRequests, ourCab, duration)
		}
		elevData.in_floor += int(elevData.direction)
		duration += TRAVEL_TIME
	}
}

func simulatedClearRequests(elevData Elevator, simRequests map[OrderType][]bool, ourCab OrderType) {
	simRequests[ourCab][elevData.in_floor] = false
	switch elevData.direction {
	case MD_Up:
		if simRequests[HALL_UP][elevData.in_floor] {
			simRequests[HALL_UP][elevData.in_floor] = false
		} else if !requestsAbove(elevData, simRequests, ourCab) {
			simRequests[HALL_DOWN][elevData.in_floor] = false
		}
	case MD_Down:
		if simRequests[HALL_DOWN][elevData.in_floor] {
			simRequests[HALL_DOWN][elevData.in_floor] = false
		} else if !requestsBelow(elevData, simRequests, ourCab) {
			simRequests[HALL_UP][elevData.in_floor] = false
		}
	default: // MD_Stop
		simRequests[HALL_DOWN][elevData.in_floor] = false
		simRequests[HALL_UP][elevData.in_floor] = false
	}
}

func requestsAbove(elevData Elevator, simRequests map[OrderType][]bool, ourCab OrderType) bool {
	for floor := elevData.in_floor + 1; floor < NUM_FLOORS; floor++ {
		if anyRequestsAtFloor(floor, simRequests, ourCab) {
			return true
		}
	}
	return false
}

func requestsBelow(elevData Elevator, simRequests map[OrderType][]bool, ourCab OrderType) bool {
	for floor := elevData.in_floor - 1; floor >= 0; floor-- {
		if anyRequestsAtFloor(floor, simRequests, ourCab) {
			return true
		}
	}
	return false
}

func anyRequests(simRequests map[OrderType][]bool, ourCab OrderType) bool {
	for floor := range NUM_FLOORS {
		if anyRequestsAtFloor(floor, simRequests, ourCab) {
			return true
		}
	}
	return false
}

func anyRequestsAtFloor(floor int, simRequests map[OrderType][]bool, ourCab OrderType) bool {
	return simRequests[HALL_DOWN][floor] || simRequests[ourCab][floor] || simRequests[HALL_UP][floor]
}

func elevShouldStop(elevData Elevator, simRequests map[OrderType][]bool, ourCab OrderType) (shouldStop bool) {
	// An out of bounds check failed here at index 4. so in_floor likley got to high
	shouldStop = false

	switch elevData.direction {
	case MD_Up:
		return (simRequests[HALL_UP][elevData.in_floor] ||
			simRequests[ourCab][elevData.in_floor] ||
			!requestsAbove(elevData, simRequests, ourCab) ||
			elevData.in_floor >= NUM_ELEVATORS-1)
	case MD_Down:
		return (simRequests[HALL_DOWN][elevData.in_floor] ||
			simRequests[ourCab][elevData.in_floor] ||
			!requestsBelow(elevData, simRequests, ourCab) ||
			elevData.in_floor == 0)
	default: // case MD_Stop
		return true
	}

}

func chooseDirection(elevData Elevator, simRequests map[OrderType][]bool, ourCab OrderType, duration int) (MotorDirection, int) {
	// check for orders in current direction of travel. if there are none, turn around
	switch elevData.direction {
	case MD_Up:
		if requestsAbove(elevData, simRequests, ourCab) {
			return MD_Up, duration
		} else if anyRequestsAtFloor(elevData.in_floor, simRequests, ourCab) {
			return MD_Stop, duration
		} else if requestsBelow(elevData, simRequests, ourCab) {
			return MD_Down, duration
		} else {
			return MD_Stop, duration
		}
	default:
		if requestsBelow(elevData, simRequests, ourCab) {
			return MD_Down, duration
		} else if anyRequestsAtFloor(elevData.in_floor, simRequests, ourCab) {
			return MD_Stop, duration
		} else if requestsAbove(elevData, simRequests, ourCab) {
			return MD_Up, duration
		} else {
			return MD_Stop, duration
		}
	}
}

func assignOrders() {

	// cab orders
	for floor := range NUM_FLOORS {
		order := ReadOrderData(CAB_FIRST+OrderType(MY_ID), floor)
		if stateFromVersionNr(order.version_nr) == ORDER_REQUESTED {
			AssignOrder(CAB_FIRST+OrderType(MY_ID), floor, 0)
		}
	}

	// hall orders
	for _, orderType := range []OrderType{HALL_UP, HALL_DOWN} {
		for floor := range NUM_FLOORS {

			order := ReadOrderData(orderType, floor)

			if stateFromVersionNr(order.version_nr) == ORDER_REQUESTED ||
				(stateFromVersionNr(order.version_nr) == ORDER_CONFIRMED &&
					time.Now().UnixMilli()-getLastProofOfWork(order.assigned_to) > ELEVATOR_TIMEOUT) {

				if stateFromVersionNr(order.version_nr) == ORDER_CONFIRMED {
					orderFailed(order.assigned_to)
					if order.assigned_to == MY_ID {
						continue
					}
				}

				cost := costFunction(orderType, floor)
				AssignOrder(orderType, floor, cost)

			} else if stateFromVersionNr(order.version_nr) == ORDER_CONFIRMED &&
				time.Now().UnixMilli()-order.assigned_at_time < BIDDING_TIME {

				cost := costFunction(orderType, floor)
				// fmt.Println("Bidding with cost", cost, "on order", orderType, floor, "against", order.assigned_cost)
				if cost+BIDDING_MIN_RAISE < order.assigned_cost {
					// fmt.Println("Got the bid with cost", cost, "on order", orderType, floor)
					AssignOrder(orderType, floor, cost)
				}
			}
		}
	}
}
