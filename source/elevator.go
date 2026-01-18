package elevator

import (
	"elevio"
	"time"
	"fmt"
)

type elev_states int

const (
	ELEV_BOOT			elev_state 	= 1
	ELEV_IDLE	 					= 2
	ELEV_RUNNING	 				= 3
	ELEV_STOP						= 4
	ELEV_DOOR_OPEN					= 5
)

type Elevator struct{
	state elev_states
	in_floor int
	ID int
	network_ID string
	direction MotorDirection
}

func (e *Elevator) Init(ID int, network_ID string){
	e.direction = MD_Stop
	elevio.SetMotorDirection(e.direction)
	e.ID = ID
	//check other elevators if they're awake.
	//go to bottom floor
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)

	e.state = ELEV_IDLE
}

func (e *Elevator) elev_open_door(){
	e.direction = MD_Stop
	elevio.SetMotorDirection(e.direction)
	elevio.SetDoorOpenLamp(true)
	time.Sleep(3 * time.Second)
	e.state = ELEV_IDLE
	elevio.SetDoorOpenLamp(false)
	//missing: transition to stop state
}

func (e *Elevator) elev_run(){
	//go through order matrix. if any eligeble order, set motor direction and continue
	//if no new orders: enter idle-mode
	//
}

func (e *Elevator) elev_stop(){
	//check if stop button is pushed in
	//once released, check if in floor. if so, enter open_door state. if not, go to nearest floor in direction it was headed and enter open door state before continuing
}

func (e *Elevator) elev_idle(){
	e.direction = MD_Stop
	elevio.SetMotorDirection(e.direction)
	//TODO: find way to get out of idle state
}

func (e *Elevator) elev_routine(){
	for{
		select{
		case e.state == ELEV_BOOT:
			e.Init(e.ID, e.network_ID)
		case e.state == ELEV_IDLE:
			e.elev_idle()
		case e.state == ELEV_DOOR_OPEN:
			e.elev_open_door()
		case e.state == ELEV_RUNNING:
			e.elev_run()
		case e.state == ELEV_STOP:
			e.elev_stop()
		}
	}
}

func (e *Elevator) get_current_floor() (int){
	return e.in_floor
}

func (e *Elevator) get_behaviour() (elev_states){
	return e.state
}

func (e *Elevator) get_direction() (MotorDirection){
	return e.direction
}