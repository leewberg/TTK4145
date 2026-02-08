package elevio

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
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
	ELEV_DOOR_OPEN
)

type Elevator struct {
	State             elev_states
	In_floor          int
	ID                int
	Direction         MotorDirection //only up or down, never stop
	Is_between_floors bool
	doorOpenTime      time.Time
}

var LocalElevator Elevator

func (e *Elevator) Elev_routine() {
	for {
		switch e.State {
		case ELEV_BOOT:
			e.Init(e.ID)
		case ELEV_IDLE:
			e.Idle()
		case ELEV_DOOR_OPEN:
			e.DoorOpen()
		case ELEV_RUNNING:
			e.Run()
		}
		//TODO: get last failed order-time. if less than 1/2s ago, enetr boot-mode
		time.Sleep(_pollRate)
	}
}

func (e *Elevator) Init(ID int) {
	e.State = ELEV_BOOT
	e.ID = ID
	e.doorOpenTime = time.Now()

	SetDoorOpenLamp(false)
	SetStopLamp(false)

	a := GetFloor()
	if a != 0 {
		SetMotorDirection(MD_Down)
		return
	}

	e.Direction = MD_Up
	SetDoorOpenLamp(false)
	SetStopLamp(false)

	e.State = ELEV_IDLE

	go e.stopRoutine()
}

func (e *Elevator) DoorOpen() {
	SetMotorDirection(MD_Stop)
	SetDoorOpenLamp(true)
	if time.Since(e.doorOpenTime) > cfg.DoorOpenTime*time.Millisecond { //doors have been open for 3+ seconds

		if e.isOrderInFloor(MDToOrdertype(e.Direction), e.In_floor) {
			db.ClearOrder(MDToOrdertype(e.Direction), e.In_floor)
		} else if e.isOrderInFloor(MDToOrdertype(e.Direction*(-1)), e.In_floor) {
			db.ClearOrder(MDToOrdertype(e.Direction*(-1)), e.In_floor)
		}

		if e.isOrderInFloor(t.GetMyCab(cfg.MyID), e.In_floor) {
			db.ClearOrder(t.GetMyCab(cfg.MyID), e.In_floor)
		}

		if !GetObstruction() { //last check before exiting door-open state
			simreq := makeSimReq(t.OrderType(2 + e.ID))
			dir, _ := ChooseDirection(*e, simreq, 10) //hva skal duration være?
			if !(dir == MD_Stop) {
				SetDoorOpenLamp(false)
				e.Direction = dir
				e.State = ELEV_RUNNING
				return
			} else {
				e.State = ELEV_IDLE
				return
			}
		}
	}
}

func (e *Elevator) Run() {
	SetMotorDirection(e.Direction)
	if e.viable_floor(e.In_floor) && !e.Is_between_floors {
		e.State = ELEV_DOOR_OPEN
		e.doorOpenTime = time.Now()
	} else {
		e.State = ELEV_IDLE

	}
}

func (e *Elevator) Idle() {
	simreq := makeSimReq(t.OrderType(2 + e.ID))
	dir, _ := ChooseDirection(*e, simreq, 10) //hva skal duration være?
	if !(dir == MD_Stop) && !GetObstruction() {
		SetDoorOpenLamp(false)
		e.Direction = dir
		e.State = ELEV_RUNNING
		return
	} else {
		e.Direction = e.Direction * (-1)
		if e.viable_floor(e.In_floor) && !GetObstruction() {
			e.State = ELEV_DOOR_OPEN
			e.doorOpenTime = time.Now()
		}
	}
	SetDoorOpenLamp(true)
	SetMotorDirection(MD_Stop)
}

func (e *Elevator) stopRoutine() bool {
	j := 0
	for i := range cfg.NumFloors {
		if !e.viable_floor(i) {
			j++
		}
	}
	return j == cfg.NumFloors
}
