package elevio

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
	"time"
)

type elevator struct {
	state             t.Elev_states
	in_floor          int
	direction         t.MotorDirection //only up or down, never stop
	is_between_floors bool
	doorOpenedTime    time.Time
}

var LocalElevator elevator

func (e *elevator) Elev_routine() {
	ticker := time.NewTicker(cfg.BroadcastPeriod * time.Millisecond)
	defer ticker.Stop()
	for {
		switch e.state {
		case t.ELEV_BOOT:
			e.Init()
		case t.ELEV_IDLE:
			e.idle()
		case t.ELEV_DOOR_OPEN:
			e.doorOpen()
		case t.ELEV_RUNNING:
			e.run()
		}
		<-ticker.C
	}
}

func (e *elevator) Init() {
	e.state = t.ELEV_BOOT
	e.doorOpenedTime = time.Now()

	setDoorOpenLamp(false)
	setStopLamp(false)

	a := GetFloor()
	if a == -1 {
		setMotorDirection(t.MD_Down)
		return
	}

	e.direction = t.MD_Up
	setDoorOpenLamp(false)
	setStopLamp(false)

	e.state = t.ELEV_IDLE
}

func (e *elevator) doorOpen() {
	setMotorDirection(t.MD_Stop)
	setDoorOpenLamp(true)
	if time.Since(e.doorOpenedTime) > cfg.DoorOpenTime*time.Millisecond {
		orderMatrix := db.GetOrderMatrix(t.GetMyCab(cfg.MyID))
		if isOrder(mdToOrdertype(e.direction), e.in_floor, orderMatrix) {
			db.ClearOrder(mdToOrdertype(e.direction), e.in_floor)
		} else if isOrder(mdToOrdertype(e.direction*(-1)), e.in_floor, orderMatrix) {
			db.ClearOrder(mdToOrdertype(e.direction*(-1)), e.in_floor)
		}

		db.ClearOrder(t.GetMyCab(cfg.MyID), e.in_floor)

		if !getObstruction() { //last check before exiting door-open state
			dir := chooseDirection(*e, orderMatrix)
			if !(dir == t.MD_Stop) {
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

func (e *elevator) run() {
	setMotorDirection(e.direction)
	if !e.is_between_floors {
		orderMatrix := db.GetOrderMatrix(t.GetMyCab(cfg.MyID))
		if isOrder(mdToOrdertype(e.direction), e.in_floor, orderMatrix) || isOrder(t.OrderType(cfg.MyID), e.in_floor, orderMatrix) {
			e.state = t.ELEV_DOOR_OPEN
			e.doorOpenedTime = time.Now()
		} else {
			e.state = t.ELEV_IDLE
		}
	}
}

func (e *elevator) idle() {
	setDoorOpenLamp(false)
	orderMatrix := db.GetOrderMatrix(t.GetMyCab(cfg.MyID))
	dir := chooseDirection(*e, orderMatrix)

	if !(dir == t.MD_Stop) {
		e.direction = dir
		e.state = t.ELEV_RUNNING
		return
	} else {
		e.direction = e.direction * (-1)
		if isOrder(mdToOrdertype(e.direction), e.in_floor, orderMatrix) || isOrder(t.OrderType(cfg.MyID), e.in_floor, orderMatrix) {
			e.state = t.ELEV_DOOR_OPEN
			e.doorOpenedTime = time.Now()
		}
	}
	setMotorDirection(t.MD_Stop)
}
