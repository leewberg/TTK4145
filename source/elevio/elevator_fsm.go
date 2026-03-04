package elevio

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
	"time"
)

type exit_type int

const (
	SAME_DIR_AV exit_type = iota
	DIFF_DIR_AV
	NO_FIND
)

type elevator struct {
	state             t.Elev_states
	in_floor          int
	id                int
	direction         MotorDirection //only up or down, never stop
	is_between_floors bool
	doorOpenTime      time.Time
}

var LocalElevator elevator

func (e *elevator) Elev_routine() {
	for {
		switch e.state {
		case t.ELEV_BOOT:
			e.Init()
		case t.ELEV_IDLE:
			e.Idle()
		case t.ELEV_DOOR_OPEN:
			e.DoorOpen()
		case t.ELEV_RUNNING:
			e.Run()
		}
		time.Sleep(_pollRate)
	}
}

func (e *elevator) Init() {
	e.state = t.ELEV_BOOT
	e.id = cfg.MyID //not really needed?
	e.doorOpenTime = time.Now()

	setDoorOpenLamp(false)
	SetStopLamp(false)

	a := GetFloor()
	if a == -1 {
		setMotorDirection(MD_Down)
		return
	}

	e.direction = MD_Up
	setDoorOpenLamp(false)
	SetStopLamp(false)

	e.state = t.ELEV_IDLE
}

func (e *elevator) DoorOpen() {
	setMotorDirection(MD_Stop)
	setDoorOpenLamp(true)
	if time.Since(e.doorOpenTime) > cfg.DoorOpenTime*time.Millisecond {
		simreq := db.GetOrderMatrix(t.GetMyCab(cfg.MyID))
		if isOrder(mdToOrdertype(e.direction), e.in_floor, simreq) {
			db.ClearOrder(mdToOrdertype(e.direction), e.in_floor)
		} else if isOrder(mdToOrdertype(e.direction*(-1)), e.in_floor, simreq) {
			db.ClearOrder(mdToOrdertype(e.direction*(-1)), e.in_floor)
		}

		db.ClearOrder(t.GetMyCab(cfg.MyID), e.in_floor)

		if !getObstruction() { //last check before exiting door-open state
			dir := chooseDirection(*e, simreq)
			if !(dir == MD_Stop) {
				setDoorOpenLamp(false)
				e.direction = dir
				e.state = t.ELEV_RUNNING
				return
			} else {
				e.state = t.ELEV_IDLE
				return
			}
		}
	}
}

func (e *elevator) Run() {
	setMotorDirection(e.direction)
	if !e.is_between_floors {
		simreq := db.GetOrderMatrix(t.GetMyCab(cfg.MyID))
		if isOrder(mdToOrdertype(e.direction), e.in_floor, simreq) || isOrder(t.OrderType(e.id), e.in_floor, simreq) {
			e.state = t.ELEV_DOOR_OPEN
			e.doorOpenTime = time.Now()
		} else {
			e.state = t.ELEV_IDLE
		}
	}
}

func (e *elevator) Idle() {
	setDoorOpenLamp(false)
	simreq := db.GetOrderMatrix(t.OrderType(2 + e.id))
	dir := chooseDirection(*e, simreq)

	if !(dir == MD_Stop) {
		e.direction = dir
		e.state = t.ELEV_RUNNING
		return
	} else {
		e.direction = e.direction * (-1)
		if isOrder(mdToOrdertype(e.direction), e.in_floor, simreq) || isOrder(t.OrderType(e.id), e.in_floor, simreq) {
			e.state = t.ELEV_DOOR_OPEN
			e.doorOpenTime = time.Now()
		}
	}
	setMotorDirection(MD_Stop)
}
