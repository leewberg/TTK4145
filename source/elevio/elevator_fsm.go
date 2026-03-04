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
	State             t.Elev_states
	In_floor          int
	ID                int
	Direction         MotorDirection //only up or down, never stop
	Is_between_floors bool
	doorOpenTime      time.Time
}

var LocalElevator elevator

func (e *elevator) Elev_routine() {
	for {
		switch e.State {
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
	e.State = t.ELEV_BOOT
	e.ID = cfg.MyID //not really needed?
	e.doorOpenTime = time.Now()

	SetDoorOpenLamp(false)
	SetStopLamp(false)

	a := GetFloor()
	if a == -1 {
		SetMotorDirection(MD_Down)
		return
	}

	e.Direction = MD_Up
	SetDoorOpenLamp(false)
	SetStopLamp(false)

	e.State = t.ELEV_IDLE
}

func (e *elevator) DoorOpen() {
	SetMotorDirection(MD_Stop)
	SetDoorOpenLamp(true)
	if time.Since(e.doorOpenTime) > cfg.DoorOpenTime*time.Millisecond {
		simreq := db.GetOrderMatrix(t.GetMyCab(cfg.MyID))
		if isOrder(mdToOrdertype(e.Direction), e.In_floor, simreq) {
			db.ClearOrder(mdToOrdertype(e.Direction), e.In_floor)
		} else if isOrder(mdToOrdertype(e.Direction*(-1)), e.In_floor, simreq) {
			db.ClearOrder(mdToOrdertype(e.Direction*(-1)), e.In_floor)
		}

		db.ClearOrder(t.GetMyCab(cfg.MyID), e.In_floor)

		if !GetObstruction() { //last check before exiting door-open state
			dir := ChooseDirection(*e, simreq)
			if !(dir == MD_Stop) {
				SetDoorOpenLamp(false)
				e.Direction = dir
				e.State = t.ELEV_RUNNING
				return
			} else {
				e.State = t.ELEV_IDLE
				return
			}
		}
	}
}

func (e *elevator) Run() {
	SetMotorDirection(e.Direction)
	if !e.Is_between_floors {
		simreq := db.GetOrderMatrix(t.GetMyCab(cfg.MyID))
		if isOrder(mdToOrdertype(e.Direction), e.In_floor, simreq) || isOrder(t.OrderType(e.ID), e.In_floor, simreq) {
			e.State = t.ELEV_DOOR_OPEN
			e.doorOpenTime = time.Now()
		} else {
			e.State = t.ELEV_IDLE
		}
	}
}

func (e *elevator) Idle() {
	SetDoorOpenLamp(false)
	simreq := db.GetOrderMatrix(t.OrderType(2 + e.ID))
	dir := ChooseDirection(*e, simreq)
	if !(dir == MD_Stop) {

		e.Direction = dir
		e.State = t.ELEV_RUNNING
		return
	} else {
		e.Direction = e.Direction * (-1)
		if isOrder(mdToOrdertype(e.Direction), e.In_floor, simreq) || isOrder(t.OrderType(e.ID), e.In_floor, simreq) {
			e.State = t.ELEV_DOOR_OPEN
			e.doorOpenTime = time.Now()
		}
	}
	SetMotorDirection(MD_Stop)
}
