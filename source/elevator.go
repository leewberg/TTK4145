package elevio

import (
	"time"
	//"fmt"
)

type elev_states int

const (
	ELEV_BOOT      elev_states = 1
	ELEV_IDLE      elev_states = 2
	ELEV_RUNNING   elev_states = 3
	ELEV_STOP      elev_states = 4
	ELEV_DOOR_OPEN elev_states = 5
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
}

func (e *Elevator) elev_run() {
	//go through order matrix. if any eligeble order, set motor direction and continue
	//if no new orders: enter idle-mode
	//
}

func (e *Elevator) elev_stop() {
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
