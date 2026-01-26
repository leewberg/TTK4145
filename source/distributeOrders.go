package elevio

import "time"

// import "fmt"

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

	// initial considerations
	switch elevData.state {
	case ELEV_DOOR_OPEN:
		duration -= DOOR_OPEN_TIME / 2
	case ELEV_RUNNING:
		duration += TRAVEL_TIME / 2
		elevData.in_floor += int(elevData.direction)
	default:
		elevData.direction = chooseDirection(elevData, simRequests, ourCab)
	}

	for {
		if elevShouldStop(elevData, simRequests, ourCab) {

			// clears all orders for the floor
			switch elevData.direction {
			case MD_Down:
				simRequests[HALL_DOWN][elevData.in_floor] = false
			case MD_Up:
				simRequests[HALL_UP][elevData.in_floor] = false
			}
			switch chooseDirection(elevData, simRequests, ourCab) {
			case MD_Down:
				simRequests[HALL_DOWN][elevData.in_floor] = false
			case MD_Up:
				simRequests[HALL_UP][elevData.in_floor] = false
			}
			simRequests[ourCab][elevData.in_floor] = false

			if !simRequests[orderType][orderFloor] {
				return duration
			}
			duration += DOOR_OPEN_TIME
		}
		elevData.direction = chooseDirection(elevData, simRequests, ourCab)
		elevData.in_floor += int(elevData.direction)
		duration += TRAVEL_TIME
	}
}

func elevShouldStop(elevData Elevator, simRequests map[OrderType][]bool, ourCab OrderType) (shouldStop bool) {
	shouldStop = false

	switch elevData.direction {
	case MD_Down:
		shouldStop = shouldStop || simRequests[HALL_DOWN][elevData.in_floor]
	case MD_Up:
		shouldStop = shouldStop || simRequests[HALL_UP][elevData.in_floor]
	}

	switch chooseDirection(elevData, simRequests, ourCab) {
	case MD_Down:
		shouldStop = shouldStop || simRequests[HALL_DOWN][elevData.in_floor]
	case MD_Up:
		shouldStop = shouldStop || simRequests[HALL_UP][elevData.in_floor]
	}
	shouldStop = shouldStop || simRequests[ourCab][elevData.in_floor]

	return shouldStop
}

func chooseDirection(elevData Elevator, simRequests map[OrderType][]bool, ourCab OrderType) MotorDirection {
	// check for orders in current direction of travel. if there are none, turn around
	if elevData.in_floor <= 0 {
		return MD_Up
	}
	if elevData.in_floor >= NUM_FLOORS-1 {
		return MD_Down
	}

	ordersBelow := false
	ordersAbove := false

	for floor := elevData.in_floor; floor > 0; floor-- {
		if simRequests[HALL_DOWN][floor] || simRequests[ourCab][floor] || simRequests[HALL_UP][floor] {
			ordersBelow = true
			break
		}
	}

	for floor := elevData.in_floor; floor < NUM_FLOORS; floor++ {
		if simRequests[HALL_DOWN][floor] || simRequests[ourCab][floor] || simRequests[HALL_UP][floor] {
			ordersAbove = true
			break
		}
	}

	if ordersBelow && elevData.direction == MD_Down {
		return MD_Down
	} else if ordersAbove && elevData.direction == MD_Up {
		return MD_Up
	} else if ordersAbove {
		return MD_Up
	} else {
		return MD_Down
	}
}

func assignOrders() {
	isElevFunctional := getFunctionalElevators()

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
			// fmt.Println(order)

			if stateFromVersionNr(order.version_nr) == ORDER_REQUESTED ||
				stateFromVersionNr(order.version_nr) == ORDER_CONFIRMED && !isElevFunctional[order.assigned_to] {

				cost := costFunction(orderType, floor)
				AssignOrder(orderType, floor, cost)

			} else if stateFromVersionNr(order.version_nr) == ORDER_CONFIRMED &&
				time.Now().UnixMilli()-order.assigned_at_time < BIDDING_TIME {

				cost := costFunction(orderType, floor)
				if cost+BIDDING_MIN_RAISE < order.assigned_cost {
					AssignOrder(orderType, floor, cost)
				}
			}
		}
	}
}
