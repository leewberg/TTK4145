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
}

func (e *Elevator) elev_open_door() {
	SetMotorDirection(MD_Stop)
	SetDoorOpenLamp(true)
	if time.Since(e.doorOpenTime) > DOOR_OPEN_TIME*time.Millisecond { //doors have been open for 3+ seconds
		if e.switched {
			e.direction = e.direction / (-1)
			e.switched = false
		}

		ClearOrder(MDToOrdertype(e.direction), e.in_floor) //clear directional order
		ClearOrder(OrderType(2+e.ID), e.in_floor)          //clear cab-order for this elevator

		if !GetObstruction() { //last check before exiting door-open state
			if e.enter_idle() {
				e.state = ELEV_IDLE
			} else {
				SetDoorOpenLamp(false)
				e.state = ELEV_RUNNING
			}
			SetDoorOpenLamp(false)
		}
	}
}

func (e *Elevator) elev_run() {
	// fmt.Println("Dir", e.direction)
	SetMotorDirection(e.direction)
	if e.viable_floor(e.in_floor) && !e.is_between_floors {
		e.state = ELEV_DOOR_OPEN
		e.doorOpenTime = time.Now()
	} else {
		if e.enter_idle() {
			e.state = ELEV_IDLE
		}
	}
}

func (e *Elevator) elev_idle() {
	SetMotorDirection(MD_Stop)
	SetDoorOpenLamp(false)
	if !e.enter_idle() {
		SetDoorOpenLamp(false)
		e.state = ELEV_RUNNING
	}
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
		time.Sleep(_pollRate)
		// fmt.Println("State", e.state)
	}
}

func (e *Elevator) viable_floor(floor int) bool {
	if e.switched {
		order_dir := ReadOrderData(MDToOrdertype((e.direction)/(-1)), floor)
		if stateFromVersionNr(order_dir.version_nr) == ORDER_CONFIRMED && order_dir.assigned_to == e.ID && time.Now().UnixMilli()-order_dir.assigned_at_time > BIDDING_TIME {
			return true
		}
	} else {
		order_dir := ReadOrderData(MDToOrdertype(e.direction), floor)
		if stateFromVersionNr(order_dir.version_nr) == ORDER_CONFIRMED && order_dir.assigned_to == e.ID && time.Now().UnixMilli()-order_dir.assigned_at_time > BIDDING_TIME {
			return true
		}
	}

	order_cab := ReadOrderData(OrderType(2+e.ID), floor)

	if stateFromVersionNr(order_cab.version_nr) == ORDER_CONFIRMED && order_cab.assigned_to == e.ID && time.Now().UnixMilli()-order_cab.assigned_at_time > BIDDING_TIME {
		//very messy, but it checks if the order is viable by first checking if the order is confirmed and is assigned to the elevator
		return true
	}
	return false
}

func (e *Elevator) enter_idle() bool {
	//checks if the elevator should enter idle-mode
	if e.check_turn() == NO_FIND {
		if e.check_turn() != NO_FIND { //only run this twice if you didn't find an avaliable order in the first instance. if you run it twice you risk messing up the resulting directions
			return false
		}
		return true
	}
	return false
}

func (e *Elevator) check_turn() exit_type {
	switch e.direction {
	case MD_Up:
		for i := e.in_floor; i < NUM_FLOORS; i++ {
			if e.viable_floor(i) {
				//if any of the floors above are viable
				e.switched = false
				e.direction = MD_Up
				return SAME_DIR_AV
			}
		}
		for i := e.in_floor; i >= 0; i-- {
			if e.viable_floor(i) {
				//if any of the floors below are viable
				e.direction = MD_Down
				e.switched = true
				return DIFF_DIR_AV
			}
		}
		e.switched = false
		e.direction = MD_Down
		return NO_FIND
	case MD_Down:
		for i := e.in_floor; i >= 0; i-- {
			if e.viable_floor(i) {
				//if any of the floors below are viable
				e.direction = MD_Down
				e.switched = false
				return SAME_DIR_AV
			}
		}
		for i := e.in_floor; i < NUM_FLOORS; i++ {
			if e.viable_floor(i) {
				//if any of the floors above are viable
				e.direction = MD_Up
				e.switched = true
				return DIFF_DIR_AV
			}
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
