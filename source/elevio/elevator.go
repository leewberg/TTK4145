package elevio

import (
	"fmt"
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
	switched          bool
	shouldStop        bool
}

var LocalElevator Elevator

func (e *Elevator) Init(ID int) {
	e.State = ELEV_BOOT
	e.ID = ID
	e.doorOpenTime = time.Now()
	e.switched = false
	e.shouldStop = false

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

func (e *Elevator) elev_open_door() {
	SetMotorDirection(MD_Stop)
	SetDoorOpenLamp(true)
	if time.Since(e.doorOpenTime) > cfg.DoorOpenTime*time.Millisecond { //doors have been open for 3+ seconds

		if e.isOrderInFloor(MDToOrdertype(e.Direction), e.In_floor) {
			db.ClearOrder(MDToOrdertype(e.Direction), e.In_floor)
		}

		if e.isOrderInFloor(t.GetMyCab(cfg.MyID), e.In_floor) {
			db.ClearOrder(t.GetMyCab(cfg.MyID), e.In_floor)
		}

		if !GetObstruction() { //last check before exiting door-open state
			if e.enter_idle() {
				e.State = ELEV_IDLE
			} else {
				SetDoorOpenLamp(false)
				e.State = ELEV_RUNNING
			}
			SetDoorOpenLamp(false)
		}
	}
}

func (e *Elevator) elev_run() {
	SetMotorDirection(e.Direction)
	if e.viable_floor(e.In_floor) && !e.Is_between_floors {
		e.State = ELEV_DOOR_OPEN
		e.doorOpenTime = time.Now()
	} else {
		if e.shouldStop {
			e.State = ELEV_IDLE
		}
	}
}

func (e *Elevator) elev_idle() {
	SetMotorDirection(MD_Stop)
	SetDoorOpenLamp(true)
	if !e.enter_idle() && !GetObstruction() {
		SetDoorOpenLamp(false)
		e.State = ELEV_RUNNING
	}
}

func (e *Elevator) Elev_routine() {
	for {
		switch e.State {
		case ELEV_BOOT:
			e.Init(e.ID)
		case ELEV_IDLE:
			e.elev_idle()
		case ELEV_DOOR_OPEN:
			e.elev_open_door()
		case ELEV_RUNNING:
			e.elev_run()
		}
		time.Sleep(_pollRate)
	}
}

func (e *Elevator) viable_floor(floor int) bool {
	if e.switched {
		return e.isOrderInFloor(t.GetMyCab(cfg.MyID), floor) || e.isOrderInFloor(MDToOrdertype(e.Direction/(-1)), floor)
	} else {
		return e.isOrderInFloor(t.GetMyCab(cfg.MyID), floor) || e.isOrderInFloor(MDToOrdertype(e.Direction), floor)
	}
}

func (e *Elevator) stopRoutine() {
	for {
		for i := range cfg.NumFloors {
			if !(e.isOrderInFloor(t.HallUp, i) || !(e.isOrderInFloor(t.HallDown, i) || !e.isOrderInFloor(t.GetMyCab(cfg.MyID), i))) {
				e.shouldStop = true
			}
			e.shouldStop = false
		}
		time.Sleep(_pollRate)
	}
}

func (e *Elevator) isOrderInFloor(dir t.OrderType, floor int) bool {
	order := db.GetOrder(dir, floor)
	return order.GetState() == t.Confirmed && order.AssignedID == e.ID && time.Now().UnixMilli()-order.AssignedTime > cfg.BiddingTime
}

func (e *Elevator) enter_idle() bool {
	//checks if the elevator should enter idle-mode

	//needed to avoid elevator switching directions back and forth if both directions would yield to e.switched == true
	if e.switched {
		e.Direction = e.Direction / (-1)
		e.switched = false
	}

	if e.check_turn() == NO_FIND {
		if e.check_turn() != NO_FIND { //only run this twice if you didn't find an avaliable order in the first instance. if you run it twice you risk messing up the resulting directions
			return false
		}
		return true
	}
	return false
}

func (e *Elevator) check_turn() exit_type {
	switch e.Direction {
	case MD_Up:
		for i := e.In_floor; i < cfg.NumFloors; i++ {
			if e.viable_floor(i) {
				//if any of the floors above are viable
				e.switched = false
				e.Direction = MD_Up
				return SAME_DIR_AV
			}
		}
		for i := e.In_floor; i >= 0; i-- {
			if e.viable_floor(i) {
				//if any of the floors below are viable
				e.Direction = MD_Down
				e.switched = true
				return DIFF_DIR_AV
			}
		}
		e.switched = false
		e.Direction = MD_Down
		return NO_FIND
	case MD_Down:
		for i := e.In_floor; i >= 0; i-- {
			if e.viable_floor(i) {
				//if any of the floors below are viable
				e.Direction = MD_Down
				e.switched = false
				return SAME_DIR_AV
			}
		}
		for i := e.In_floor; i < cfg.NumFloors; i++ {
			if e.viable_floor(i) {
				//if any of the floors above are viable
				e.Direction = MD_Up
				e.switched = true
				return DIFF_DIR_AV
			}
		}
		e.Direction = MD_Up
		e.switched = false
		return NO_FIND
	}
	fmt.Printf("something went wrong, and we didn't register either up or down direction for elevator. \n")
	return NO_FIND
}

func MDToOrdertype(dir MotorDirection) t.OrderType {
	switch dir {
	case MD_Up:
		return t.HallUp
	case MD_Down:
		return t.HallDown
	}
	return 0
}
