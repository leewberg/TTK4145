package elevio

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
	"time"
)

type elevator struct {
	state           t.ElevStates
	inFloor         int
	direction       t.MotorDirection //only up or down, never stop
	isBetweenFloors bool
	doorOpenedTime  time.Time
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

	if !e.shouldCloseDoor() {
		return
	}

	e.completeOrders()
	e.exitFromDoorOpen()
}

func (e *elevator) shouldCloseDoor() bool {
	if getObstruction() {
		e.doorOpenedTime = time.Now()
		return false
	}
	if time.Since(e.doorOpenedTime) < cfg.DoorOpenTime*time.Millisecond {
		return false
	}
	return true
}

func (e *elevator) completeOrders() {
	orderMatrix := db.GetOrderMatrix(t.GetMyCab(cfg.MyID))

	dirOrder := mdToOrdertype(e.direction)
	revDirOrder := mdToOrdertype(e.direction * -1)

	if isOrder(dirOrder, e.inFloor, orderMatrix) {
		db.ClearOrder(dirOrder, e.inFloor)
	} else if isOrder(revDirOrder, e.inFloor, orderMatrix) {
		db.ClearOrder(revDirOrder, e.inFloor)
	}
	db.ClearOrder(t.GetMyCab(cfg.MyID), e.inFloor)
}

func (e *elevator) exitFromDoorOpen() {
	orderMatrix := db.GetOrderMatrix(t.GetMyCab(cfg.MyID))
	dir := chooseDirection(*e, orderMatrix)

	setDoorOpenLamp(false)

	if dir == t.MD_Stop {
		e.state = t.ELEV_IDLE
	} else {
		e.direction = dir
		e.state = t.ELEV_RUNNING
	}
}

func (e *elevator) run() {
	setMotorDirection(e.direction)
	if !e.isBetweenFloors {
		orderMatrix := db.GetOrderMatrix(t.GetMyCab(cfg.MyID))
		if isOrder(mdToOrdertype(e.direction), e.inFloor, orderMatrix) || isOrder(t.OrderType(cfg.MyID), e.inFloor, orderMatrix) {
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
		if isOrder(mdToOrdertype(e.direction), e.inFloor, orderMatrix) || isOrder(t.OrderType(cfg.MyID), e.inFloor, orderMatrix) {
			e.state = t.ELEV_DOOR_OPEN
			e.doorOpenedTime = time.Now()
		}
	}
	setMotorDirection(t.MD_Stop)
}
