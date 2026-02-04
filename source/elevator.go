package elevio

import (
	"fmt"
	"time"
)

type elev_states int
type exit_type int

const (
	SAME_DIR_AV exit_type = iota
	DIFF_DIR_AV
	NO_FIND
)

const (
	ELEV_BOOT elev_states = iota
	ELEV_IDLE
	ELEV_RUNNING
	ELEV_DOOR_OPEN
)

type Elevator struct {
	state             elev_states
	in_floor          int
	ID                int
	direction         MotorDirection //only up or down, never stop
	is_between_floors bool
	doorOpenTime      time.Time
	switched          bool
}

func (e *Elevator) Init(ID int) {
	e.state = ELEV_BOOT
	e.ID = ID
	e.doorOpenTime = time.Now()
	e.switched = false

	SetDoorOpenLamp(false)
	SetStopLamp(false)

	a := GetFloor()
	if a != 0 {
		SetMotorDirection(MD_Down)
		return
	}

	e.direction = MD_Up
	SetDoorOpenLamp(false)
	SetStopLamp(false)

	e.state = ELEV_IDLE

	go e.stopRoutine()
}

func (e *Elevator) elev_open_door() {
	SetMotorDirection(MD_Stop)
	SetDoorOpenLamp(true)
	if time.Since(e.doorOpenTime) > DOOR_OPEN_TIME*time.Millisecond { //doors have been open for 3+ seconds

		if e.isOrderInFloor(MDToOrdertype(e.direction), e.in_floor) {
			ClearOrder(MDToOrdertype(e.direction), e.in_floor)
		} else if e.isOrderInFloor(MDToOrdertype(e.direction*(-1)), e.in_floor) { //if we're only on an up-order, direction down, the doors shouldn't have to be open for 6 seconds, because that'd be dumb
			ClearOrder(MDToOrdertype(e.direction*(-1)), e.in_floor)
		}

		if e.isOrderInFloor(OrderType(2+e.ID), e.in_floor) {
			ClearOrder(OrderType(2+e.ID), e.in_floor)
		}

		if !GetObstruction() { //last check before exiting door-open state
			simreq := makeSimReq(OrderType(2 + e.ID))
			dir, _ := chooseDirection(*e, simreq, CAB_FIRST+OrderType(e.ID), 10) //hva skal duration være?
			if !(dir == MD_Stop) {
				SetDoorOpenLamp(false)
				e.direction = dir
				e.state = ELEV_RUNNING
				return
			} else {
				e.state = ELEV_IDLE
				return
			}
		}
	}
}

func (e *Elevator) elev_run() {
	SetMotorDirection(e.direction)
	if e.viable_floor(e.in_floor) && !e.is_between_floors {
		e.state = ELEV_DOOR_OPEN
		e.doorOpenTime = time.Now()
	} else {
		e.state = ELEV_IDLE

	}
}

func makeSimReq(ourCab OrderType) map[OrderType][]bool {
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
	return simRequests
}

func (e *Elevator) elev_idle() {
	simreq := makeSimReq(OrderType(2 + e.ID))
	dir, _ := chooseDirection(*e, simreq, CAB_FIRST+OrderType(e.ID), 10) //hva skal duration være?
	if !(dir == MD_Stop) && !GetObstruction() {
		SetDoorOpenLamp(false)
		e.direction = dir
		e.state = ELEV_RUNNING
		return
	} else {
		e.direction = e.direction * (-1)
		if e.viable_floor(e.in_floor) && !GetObstruction() {
			e.state = ELEV_DOOR_OPEN
			e.doorOpenTime = time.Now()
		}
	}
	SetDoorOpenLamp(true)
	SetMotorDirection(MD_Stop)
}

func (e *Elevator) Elev_routine() {
	for {
		switch e.state {
		case ELEV_BOOT:
			e.Init(e.ID)
		case ELEV_IDLE:
			e.elev_idle()
		case ELEV_DOOR_OPEN:
			e.elev_open_door()
		case ELEV_RUNNING:
			e.elev_run()
		}
		//TODO: get last failed order-time. if less than 1/2s ago, enetr boot-mode
		time.Sleep(_pollRate)
	}
}

func (e *Elevator) viable_floor(floor int) bool {
	/*if e.switched {
		return e.isOrderInFloor(OrderType(2+e.ID), floor) || e.isOrderInFloor(MDToOrdertype(e.direction/(-1)), floor)
	} else {
		return e.isOrderInFloor(OrderType(2+e.ID), floor) || e.isOrderInFloor(MDToOrdertype(e.direction), floor)
	}*/
	return e.isOrderInFloor(OrderType(2+e.ID), floor) || e.isOrderInFloor(MDToOrdertype(e.direction), floor) //|| e.isOrderInFloor(MDToOrdertype(e.direction/(-1)), floor)
}

func (e *Elevator) stopRoutine() bool {
	j := 0
	for i := range NUM_FLOORS {
		if !e.viable_floor(i) {
			j++
		}
	}
	return j == NUM_FLOORS
}

func (e *Elevator) isOrderInFloor(dir OrderType, floor int) bool {
	order := ReadOrderData(dir, floor)
	return stateFromVersionNr(order.version_nr) == ORDER_CONFIRMED && order.assigned_to == e.ID && time.Now().UnixMilli()-order.assigned_at_time > BIDDING_TIME
}

func (e *Elevator) enter_idle() bool {
	//checks if the elevator should enter idle-mode

	//needed to avoid elevator switching directions back and forth if both directions would yield to e.switched == true
	if e.switched {
		e.direction = e.direction / (-1)
		e.switched = false
	}

	if e.check_turn() == NO_FIND {
		if e.check_turn() != NO_FIND { //only run this twice if you didn't find an avaliable order in the first instance. if you run it twice you risk messing up the resulting directions
			return false
		}
		return true
	}
	return false
}

func (e *Elevator) ordersAbove() bool {
	for i := e.in_floor; i < NUM_FLOORS; i++ {
		if e.viable_floor(i) {
			return true
		}
	}
	return false
}

func (e *Elevator) ordersBelow() bool {
	for i := e.in_floor; i >= 0; i-- {
		if e.viable_floor(i) {
			return true
		}
	}
	return false
}

func (e *Elevator) check_turn() exit_type {
	switch e.direction {
	case MD_Up:
		if e.ordersAbove() {
			e.switched = false
			e.direction = MD_Up
			return SAME_DIR_AV
		}
		if e.ordersBelow() {
			e.direction = MD_Down
			e.switched = true
			return DIFF_DIR_AV
		}
		e.switched = false
		e.direction = MD_Down
		return NO_FIND
	case MD_Down:
		if e.ordersBelow() {
			e.direction = MD_Down
			e.switched = false
			return SAME_DIR_AV
		}
		if e.ordersAbove() {
			e.direction = MD_Up
			e.switched = true
			return DIFF_DIR_AV
		}
		e.direction = MD_Up
		e.switched = false
		return NO_FIND
	}
	fmt.Printf("something went wrong, and we didn't register either up or down direction for elevator. \n")
	return NO_FIND
}

func MDToOrdertype(dir MotorDirection) OrderType {
	switch dir {
	case MD_Up:
		return HALL_UP
	case MD_Down:
		return HALL_DOWN
	}
	return 0
}
