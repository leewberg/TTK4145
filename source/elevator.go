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
	ELEV_STOP
	ELEV_DOOR_OPEN
)

type Elevator struct {
	state             elev_states
	in_floor          int
	ID                int
	network_ID        string
	direction         MotorDirection //only up or down, never stop
	initialized       bool
	justStopped       bool
	is_between_floors bool
	doorOpenTime      time.Time
	switched          bool
}

func (e *Elevator) Init(ID int, network_ID string) {
	e.state = ELEV_BOOT
	e.ID = ID
	e.network_ID = network_ID
	e.doorOpenTime = time.Now()
	e.switched = false

	SetDoorOpenLamp(false)
	SetStopLamp(false)

	a := GetFloor()
	for a != 0 {
		SetMotorDirection(MD_Down)
		a = GetFloor()
	}

	e.direction = MD_Up
	SetDoorOpenLamp(false)
	SetStopLamp(false)
	declareElevatorFunctional()

	e.state = ELEV_IDLE
}

func (e *Elevator) elev_open_door() {
	SetMotorDirection(MD_Stop)
	SetDoorOpenLamp(true)
	if time.Since(e.doorOpenTime) > 3*time.Second { //doors have been open for 3+ seconds
		if e.switched {
			ClearOrder(MDToOrdertype((e.direction / (-1))), e.in_floor)
		} else {
			ClearOrder(MDToOrdertype(e.direction), e.in_floor) //clear directional order
		}
		ClearOrder(OrderType(2+e.ID), e.in_floor) //clear cab-order for this elevator

		//check if enter idle mode: run check turn twice. if the we don't find any avaliable orders above or below, we enter idle mode. only run second check if first check has no finds, as running the check twice will mess up the direction
		if !GetObstruction() { //last check before exiting door-open state
			if e.switched {
				e.direction = e.direction / (-1)
				e.switched = false
			}
			if e.check_turn() == NO_FIND {
				if e.check_turn() != NO_FIND {
					SetDoorOpenLamp(false)
					e.state = ELEV_RUNNING
				} else {
					e.state = ELEV_IDLE
				}
			} else {
				SetDoorOpenLamp(false)
				e.state = ELEV_RUNNING
			}
			declareElevatorFunctional()
			SetDoorOpenLamp(false)
		}
	}
}

func (e *Elevator) elev_run() {
	// TODO: Make sure this stops if we no longer have an order to go after. Someone could have stolen the order :)
	SetMotorDirection(e.direction)
	declareElevatorFunctional()
	if e.viable_floor(e.in_floor) && !e.is_between_floors {
		e.state = ELEV_DOOR_OPEN
		e.doorOpenTime = time.Now()
	} /*else {
		if e.check_turn() == NO_FIND && e.check_turn() == NO_FIND {
			e.state = ELEV_IDLE
		}
	}*/
}

func (e *Elevator) elev_stop() {
	SetStopLamp(true)
	if e.justStopped {
		a := GetFloor()
		if a == -1 {
			for a == -1 {
				SetMotorDirection(e.direction)
				a = GetFloor()
			}
			e.state = ELEV_DOOR_OPEN
			e.justStopped = false
		} else {
			e.state = ELEV_DOOR_OPEN
			e.justStopped = false
		}
		SetStopLamp(false)
	}
}

func (e *Elevator) elev_idle() {
	SetMotorDirection(MD_Stop)
	SetDoorOpenLamp(true)
	if e.check_turn() == NO_FIND {
		if e.check_turn() != NO_FIND {
			SetDoorOpenLamp(false)
			e.state = ELEV_RUNNING
		}
	} else {
		SetDoorOpenLamp(false)
		e.state = ELEV_RUNNING
	}
	declareElevatorFunctional()
}

func (e *Elevator) Elev_routine() {
	for {
		switch e.state {
		case ELEV_BOOT:
			e.Init(e.ID, e.network_ID)
		case ELEV_IDLE:
			e.elev_idle()
		case ELEV_DOOR_OPEN:
			e.elev_open_door()
		case ELEV_RUNNING:
			e.elev_run()
		case ELEV_STOP:
			e.elev_stop()
		}
		time.Sleep(_pollRate)
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

func (e *Elevator) check_turn() exit_type {
	//returns bool based on if there's no viable floors above or below
	switch e.direction {
	case MD_Up:
		for i := e.in_floor; i < NUM_FLOORS; i++ {
			if e.viable_floor(i) {
				//if any of the floors above are viable
				e.switched = false
				//e.direction = MD_Up
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
