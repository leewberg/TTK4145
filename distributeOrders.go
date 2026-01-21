package main

func costFunction(orderType OrderType, orderFloor int, elevID int) int {
	// finds the cost for elevator elevID to do a spesific order, by simulating execution
	elevData := getElevData(elevID)
	duration := 0
	ourCab := OrderType(2 + elevID)

	// copy down data so we don't override the actual orders
	var simRequests map[OrderType][]bool
	for _, orderType := range []OrderType{HALL_UP, HALL_DOWN, ourCab} {
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
	ordersBelow := false
	ordersAbove := false

	for floor := elevData.last_floor; floor > 0; floor-- {
		if simRequests[HALL_DOWN][floor] || simRequests[ourCab][floor] || simRequests[HALL_UP][floor] {
			ordersBelow = true
			break
		}
	}

	for floor := elevData.last_floor; floor > 0; floor-- {
		if simRequests[HALL_DOWN][floor] || simRequests[ourCab][floor] || simRequests[HALL_UP][floor] {
			ordersAbove = true
			break
		}
	}

	if ordersBelow || elevData.direction == DIR_DOWN {
		return DIR_DOWN
	} else if ordersAbove || elevData.direction == DIR_UP {
		return DIR_UP
	} else if ordersAbove {
		return DIR_UP
	} else {
		return DIR_DOWN
	}
}

func findBestElevForOrder(orderType OrderType, orderFloor int, isElevFunctional []bool) (bestElev int, bestCost int) {
	bestElev = 0
	bestCost = INF

	for elevID := range NUM_ELEVATORS {
		// this automatically enforces the rule that lower IDs have tiebreak priority
		if isElevFunctional[elevID] {
			cost := costFunction(orderType, orderFloor, elevID)
			if cost < bestCost {
				bestCost = cost
				bestElev = elevID
			}
		}
	}
	return bestElev, bestCost
}

func assignOrders() {
	// iterate through every order, check if it needs assignment and find cost

	isElevFunctional := getFunctionalElevators()

	// cab orders
	for orderType := CAB_1; orderType < NUM_ELEVATORS+2; orderType++ {
		for floor := range NUM_FLOORS {
			// ordertype - 2 is the elevatorID corresponding to the cab request
			assignOrder(orderType, floor, int(orderType-2), 0)
		}
	}

	// hall orders
	for _, orderType := range []OrderType{HALL_UP, HALL_DOWN} {
		for floor := range NUM_FLOORS {

			order := readOrderData(orderType, floor)
			if stateFromVersionNr(order.version_nr) == ORDER_REQUESTED ||
				!isElevFunctional[order.assigned_to] {

				bestElev, cost := findBestElevForOrder(orderType, floor, isElevFunctional)
				assignOrder(orderType, floor, bestElev, cost)
			}
		}
	}
}
