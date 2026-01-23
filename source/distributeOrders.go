package elevio

import "time"

func costFunction(orderType OrderType, orderFloor int, elevID int) int {
	// finds the cost for elevator elevID to do a spesific order, by simulating execution
	elevData := getElevData(elevID)
	duration := 0
	ourCab := OrderType(2 + elevID)

	if elevData.last_floor == -1 {
		return INF // elevator not initialized
	}

	// copy down data so we don't override the actual orders
	simRequests := make(map[OrderType][]bool)
	for _, orderType := range []OrderType{HALL_UP, HALL_DOWN, ourCab} {
		simRequests[orderType] = make([]bool, NUM_FLOORS)
		for floor := range NUM_FLOORS {
			orderData := readOrderData(orderType, floor)
			if stateFromVersionNr(orderData.version_nr) == ORDER_CONFIRMED &&
				orderData.assigned_to == elevID {
				simRequests[orderType][floor] = true
			}
		}
	}
	simRequests[orderType][orderFloor] = true

	// initial considerations
	switch elevData.state {
	case STATE_DOOR_OPEN:
		duration -= DOOR_OPEN_TIME / 2
	case STATE_RUNNING:
		duration += TRAVEL_TIME / 2
		elevData.last_floor += int(elevData.direction)
	default:
		elevData.direction = chooseDirection(elevData, simRequests, ourCab)
	}

	for {
		if elevShouldStop(elevData, simRequests, ourCab) {

			// clears all orders for the floor
			switch elevData.direction {
			case DIR_DOWN:
				simRequests[HALL_DOWN][elevData.last_floor] = false
			case DIR_UP:
				simRequests[HALL_UP][elevData.last_floor] = false
			}
			switch chooseDirection(elevData, simRequests, ourCab) {
			case DIR_DOWN:
				simRequests[HALL_DOWN][elevData.last_floor] = false
			case DIR_UP:
				simRequests[HALL_UP][elevData.last_floor] = false
			}
			simRequests[ourCab][elevData.last_floor] = false

			if !simRequests[orderType][orderFloor] {
				return duration
			}
			duration += DOOR_OPEN_TIME
		}
		elevData.direction = chooseDirection(elevData, simRequests, ourCab)
		elevData.last_floor += int(elevData.direction)
		duration += TRAVEL_TIME
	}
}

func elevShouldStop(elevData ElevatorData, simRequests map[OrderType][]bool, ourCab OrderType) (shouldStop bool) {
	shouldStop = false

	switch elevData.direction {
	case DIR_DOWN:
		shouldStop = shouldStop || simRequests[HALL_DOWN][elevData.last_floor]
	case DIR_UP:
		shouldStop = shouldStop || simRequests[HALL_UP][elevData.last_floor]
	}

	switch chooseDirection(elevData, simRequests, ourCab) {
	case DIR_DOWN:
		shouldStop = shouldStop || simRequests[HALL_DOWN][elevData.last_floor]
	case DIR_UP:
		shouldStop = shouldStop || simRequests[HALL_UP][elevData.last_floor]
	}
	shouldStop = shouldStop || simRequests[ourCab][elevData.last_floor]

	return shouldStop
}

func chooseDirection(elevData ElevatorData, simRequests map[OrderType][]bool, ourCab OrderType) Direction {
	// check for orders in current direction of travel. if there are none, turn around
	if elevData.last_floor <= 0 {
		return DIR_UP
	}
	if elevData.last_floor >= NUM_FLOORS-1 {
		return DIR_DOWN
	}

	ordersBelow := false
	ordersAbove := false

	for floor := elevData.last_floor; floor > 0; floor-- {
		if simRequests[HALL_DOWN][floor] || simRequests[ourCab][floor] || simRequests[HALL_UP][floor] {
			ordersBelow = true
			break
		}
	}

	for floor := elevData.last_floor; floor < NUM_FLOORS; floor++ {
		if simRequests[HALL_DOWN][floor] || simRequests[ourCab][floor] || simRequests[HALL_UP][floor] {
			ordersAbove = true
			break
		}
	}

	if ordersBelow && elevData.direction == DIR_DOWN {
		return DIR_DOWN
	} else if ordersAbove && elevData.direction == DIR_UP {
		return DIR_UP
	} else if ordersAbove {
		return DIR_UP
	} else {
		return DIR_DOWN
	}
}

func assignOrders() {
	isElevFunctional := getFunctionalElevators()

	// cab orders
	for floor := range NUM_FLOORS {
		assignOrder(CAB_FIRST+MY_ID, floor, 0)
	}

	// hall orders
	for _, orderType := range []OrderType{HALL_UP, HALL_DOWN} {
		for floor := range NUM_FLOORS {

			order := readOrderData(orderType, floor)

			if stateFromVersionNr(order.version_nr) == ORDER_REQUESTED ||
				stateFromVersionNr(order.version_nr) == ORDER_CONFIRMED && !isElevFunctional[order.assigned_to] {

				cost := costFunction(orderType, floor, MY_ID)
				assignOrder(orderType, floor, cost)

			} else if stateFromVersionNr(order.version_nr) == ORDER_CONFIRMED &&
				time.Now().UnixMilli()-order.assigned_at_time < BIDDING_TIME {

				cost := costFunction(orderType, floor, MY_ID)
				if cost+BIDDING_MIN_RAISE < order.assigned_cost {
					assignOrder(orderType, floor, cost)
				}
			}
		}
	}
}
