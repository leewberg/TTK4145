package elevio

import (
	"time"
	//"fmt"
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
	state       elev_states
	in_floor    int
	ID          int
	network_ID  string
	direction   MotorDirection
	initialized bool
	obstacle    bool
	justStopped bool
}

func (e *Elevator) Init(ID int, network_ID string) {
	SetMotorDirection(e.direction)
	e.ID = ID
	e.network_ID = network_ID
	//check other elevators if they're awake.
	//this can be a for loop, so that all elevators are always checking if they need to be initialized or not. this allows for an elevator to die and come back again (wow, jesus parallell) without the system being rebooted
	//go to bottom floor
	SetDoorOpenLamp(false)
	SetStopLamp(false)

	e.state = ELEV_IDLE
}

func (e *Elevator) elev_open_door() {
	e.direction = MD_Stop
	SetMotorDirection(e.direction)
	SetDoorOpenLamp(true)
	time.Sleep(3 * time.Second)
	e.state = ELEV_IDLE
	SetDoorOpenLamp(false)
	//missing: turn off order lights (cab and direction)
}

func (e *Elevator) elev_run() {
	//go through order matrix. if any eligeble order, set motor direction and continue
	//if no new orders: enter idle-mode
	//update e.in_floor
}

func (e *Elevator) elev_stop() {
	SetStopLamp(true)
	if e.justStopped {
		a := <-drv_floors
		if a == -1 {
			for a == -1 {
				SetMotorDirection(e.direction)
				a = <-drv_floors
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
	e.direction = MD_Stop
	SetMotorDirection(e.direction)
	//TODO: find way to get out of idle state
}

func (e *Elevator) elev_routine() {
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
	}
}

func (e *Elevator) hability_rouitine() {
	//checks if elevator is alive every x seconds

}

func (e *Elevator) viable_floor(floor int) bool {
	order_dir := readOrderData((OrderType(e.direction)), floor)
	order_cab := readOrderData(OrderType(1+e.ID), floor)

	if (stateFromVersionNr(order_dir.version_nr) == ORDER_CONFIRMED && order_dir.assigned_to == e.ID) || (stateFromVersionNr(order_cab.version_nr) == ORDER_CONFIRMED && order_cab.assigned_to == e.ID) {
		//very messy, but it checks if the order is viable by first checking if the order is confirmed and is assigned to the elevator
		return true
	}
	return false
}

func (e *Elevator) check_turn() {
	switch e.direction {
	case MD_Up:
		for i := e.in_floor; i < NUM_FLOORS; i++ {
			if e.viable_floor(i) {
				//if any of the floors above are viable
				break
			}
		}
		e.direction = MD_Down
	case MD_Down:
		for i := e.in_floor; i >= 0; i-- {
			if e.viable_floor(i) {
				//if any of the floors below are viable
				break
			}
		}
		e.direction = MD_Up
	}
}

func (e *Elevator) get_current_floor() int {
	return e.in_floor
}

func (e *Elevator) get_behaviour() elev_states {
	return e.state
}

func (e *Elevator) get_direction() MotorDirection {
	return e.direction
}

/*
pending functions:
	func (e *Elevator) enable_stop() redundant
		seperate routine that checks if the stop button is pushed in or not. if it is, the elevator immediately transitions to the stop-state

	func (e *Elevator) send_order()
		whenever an elevator detects an order, it'll send it to the allOrdersData matrix with floor and ordertype-.info. sends also own ID
		|| whoopsie can use request

	func (e *Elevator) send_update()
		sends update of worldview into the void. need network module to finish / know how to struct

notes to self
	allOrdersData[ordeType][orderFloor]

	if we get performance issues wrt. checking the performance matrix while in the floor, i can add an extra class atribute which tells the elevator if it's supposed to stop in the next floor or not.
*/
