package elevio

import (
	"fmt"
	"time"
)

type elev_states int

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
	obstacle          bool
	justStopped       bool
	is_between_floors bool
}

func (e *Elevator) Init(ID int, network_ID string) {
	e.state = ELEV_BOOT
	e.ID = ID
	e.network_ID = network_ID
	//maybe more network init needed, idk

	SetDoorOpenLamp(false)
	SetStopLamp(false)

	a := GetFloor()
	for a != 0 {
		//go to bottom floor (maybe not needed, but was req for previous elevator lab)
		SetMotorDirection(MD_Down)
		a = GetFloor()
	}

	e.direction = MD_Up
	SetDoorOpenLamp(false)
	SetStopLamp(false)

	e.state = ELEV_IDLE
}

func (e *Elevator) elev_open_door() {
	SetMotorDirection(MD_Stop)
	SetDoorOpenLamp(true)
	if e.obstacle {
		for e.obstacle == true {
			e.obstacle = GetObstruction()
		}
	}
	time.Sleep(3 * time.Second)
	ClearOrder(MDToOrdertype(e.direction), e.in_floor) //clear directional order
	ClearOrder(OrderType(2+e.ID), e.in_floor)          //clear cab-order for this elevator
	//check if enter idle mode: run check turn twice. if the direction is the same after two turns (meaning there's no viable orders below or above), we enter idle mode. if there's none above but there are below, the direction will only be flipped once
	t1 := e.check_turn()
	t2 := e.check_turn()
	if t1 && t2 { //if direction was flipped twice
		//no viable orders above or below
		e.state = ELEV_IDLE
	} else {
		e.state = ELEV_RUNNING
	}

	SetDoorOpenLamp(false)

}

func (e *Elevator) elev_run() {
	// TODO: Make sure this stops if we no longer have an order to go after. Someone could have stolen the order :)
	SetMotorDirection(e.direction)
	if e.viable_floor(e.in_floor) && !e.is_between_floors {
		e.state = ELEV_DOOR_OPEN
	}
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
	//turn off all lights ? see task specification when it comes out later
}

func (e *Elevator) elev_idle() {
	SetMotorDirection(MD_Stop)
	SetDoorOpenLamp(true)
	t1 := e.check_turn()
	t2 := e.check_turn()
	if !(t1 && t2) { //viable order detected!
		//enter open-door mode to ensure doors stay open for 3 more seconds. after this, we will enter running mode
		e.state = ELEV_RUNNING
		SetDoorOpenLamp(false)
	}
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
		// fmt.Println("State:", e.state)
	}
}

func (e *Elevator) Hability_routine() {
	for {
		declareElevatorFunctional() //may need to send in elevator ID here
		time.Sleep(100 * time.Millisecond)
	}
}

func (e *Elevator) viable_floor(floor int) bool {
	order_dir := ReadOrderData(MDToOrdertype(e.direction), floor)
	order_cab := ReadOrderData(OrderType(2+e.ID), floor)

	if (stateFromVersionNr(order_dir.version_nr) == ORDER_CONFIRMED && order_dir.assigned_to == e.ID && time.Now().UnixMilli()-order_dir.assigned_at_time > BIDDING_TIME) || (stateFromVersionNr(order_cab.version_nr) == ORDER_CONFIRMED && order_cab.assigned_to == e.ID) {
		//very messy, but it checks if the order is viable by first checking if the order is confirmed and is assigned to the elevator
		return true
	}
	return false
}

func (e *Elevator) check_turn() bool {
	//returns bool based on if direction was flipped or not
	switch e.direction {
	case MD_Up:
		for i := e.in_floor; i < NUM_FLOORS; i++ {
			if e.viable_floor(i) {
				//if any of the floors above are viable
				return false
			}
		}
		e.direction = MD_Down
		return true
	case MD_Down:
		for i := e.in_floor; i >= 0; i-- {
			if e.viable_floor(i) {
				//if any of the floors below are viable
				return false
			}
		}
		e.direction = MD_Up
		return true
	}
	fmt.Printf("something went wrong, and we didn't register either up or down direction for elevator. \n")
	return false
}

func (e *Elevator) get_current_floor() int {
	return e.in_floor
}

func (e *Elevator) get_behaviour() elev_states {
	return e.state
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
