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
	state      elev_states
	in_floor   int
	ID         int
	network_ID string
	direction  MotorDirection
}

func (e *Elevator) Init(ID int, network_ID string) {
	e.direction = MD_Stop
	SetMotorDirection(e.direction)
	e.ID = ID
	//check other elevators if they're awake.
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
	//missing: transition to stop state
	//missing: turn off order lights (cab and direction)
}

func (e *Elevator) elev_run() {
	//go through order matrix. if any eligeble order, set motor direction and continue
	//if no new orders: enter idle-mode
	//
}

func (e *Elevator) elev_stop() {
	SetStopLamp(true)
	//check if stop button is pushed in
	//once released, check if in floor. if so, enter open_door state. if not, go to nearest floor in direction it was headed and enter open door state before continuing
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
	func (e *Elevator) viable_floor() bool
		takes in elevator object and it's current floor, and checks it against the allOrders-matrix (as defined in database module), to see if the current floor either has an up-order and/or a cab-order. if it does, it enters the door-open state

	func (e *Elevator) check_turn()
		checks the remaining floors in it's direction (so all floors above it if it's going up, all below if down.) in the larger allOrdersData matrix. if it sees that there's no viable orders above it (or below it. whatever), it'll turn. if not, it continues in stated direction
		can use a for-loop and viable_floor() function for this
		can use function readOrderData(orderType, orderFloor) for this. it returns an orderData object with attribute assigned_to, which tells if its assigned to us or not
		can also use orderVersion2State(order_version_nr int) OrderState for this. tells us if it's [CLEAR, REQUESTED, CONFIRMED]

	func (e *Elevator) enable_stop()
		seperate routine that checks if the stop button is pushed in or not. if it is, the elevator immediately transitions to the stop-state

	func (e *Elevator) send_order()
		whenever an elevator detects an order, it'll send it to the allOrdersData matrix with floor and ordertype-.info. sends also own ID
		|| whoopsie can use request

	func (e *Elevator) clear_order()
		whenever an elevator finishes an order, it sends a message to the data-matrix that the order is finished before turning off the lights || whoopsie already made. from database-module: clearOrder(orderType OrderType, orderFloor OrderType)

	func (e *Elevator) send_update()
		sends update of worldview into the void. need network module to finish / know how to struct

notes to self
	allOrdersData[ordeType][orderFloor]

	if we get performance issues wrt. checking the performance matrix while in the floor, i can add an extra class atribute which tells the elevator if it's supposed to stop in the next floor or not.
*/
