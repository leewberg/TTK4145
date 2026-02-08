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

func (e *Elevator) elev_open_door() {
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
			dir, _ := chooseDirection(*e, simreq, 10) //hva skal duration være?
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

func (e *Elevator) elev_run() {
	SetMotorDirection(e.Direction)
	if e.viable_floor(e.In_floor) && !e.Is_between_floors {
		e.State = ELEV_DOOR_OPEN
		e.doorOpenTime = time.Now()
	} else {
		e.State = ELEV_IDLE

	}
}

func makeSimReq(ourCab t.OrderType) map[t.OrderType][]bool {
	simRequests := make(map[t.OrderType][]bool)
	for _, orderType := range []t.OrderType{t.HallUp, t.HallDown, ourCab} {
		simRequests[orderType] = make([]bool, cfg.NumFloors)
		for floor := range cfg.NumFloors {
			orderData := db.GetOrder(orderType, floor)
			if orderData.GetState() == t.Confirmed &&
				t.OrderType(orderData.AssignedID) == ourCab { //need to make sure that switching to ourCab was viable desicion
				simRequests[orderType][floor] = true
			}
		}
	}
	return simRequests
}

func (e *Elevator) elev_idle() {
	simreq := makeSimReq(t.OrderType(2 + e.ID))
	dir, _ := chooseDirection(*e, simreq, 10) //hva skal duration være?
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
		//TODO: get last failed order-time. if less than 1/2s ago, enetr boot-mode
		time.Sleep(_pollRate)
	}
}

func (e *Elevator) viable_floor(floor int) bool {

	return e.isOrderInFloor(t.OrderType(2+e.ID), floor) || e.isOrderInFloor(MDToOrdertype(e.Direction), floor)
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

func (e *Elevator) isOrderInFloor(dir t.OrderType, floor int) bool {
	order := db.GetOrder(dir, floor)
	return order.GetState() == t.Confirmed && order.AssignedID == e.ID && time.Now().UnixMilli()-order.AssignedTime > cfg.BiddingTime
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

func chooseDirection(elevData Elevator, simRequests map[t.OrderType][]bool, duration int) (MotorDirection, int) {
	// check for orders in current direction of travel. if there are none, turn around
	switch elevData.Direction {
	case MD_Up:
		if requestsAbove(elevData, simRequests) {
			return MD_Up, duration
		} else if anyRequestsAtFloor(elevData.In_floor, simRequests) {
			return MD_Stop, duration
		} else if requestsBelow(elevData, simRequests) {
			return MD_Down, duration + 5000
		} else {
			return MD_Stop, duration
		}
	default:
		if requestsBelow(elevData, simRequests) {
			return MD_Down, duration
		} else if anyRequestsAtFloor(elevData.In_floor, simRequests) {
			return MD_Stop, duration
		} else if requestsAbove(elevData, simRequests) {
			return MD_Up, duration + 5000
		} else {
			return MD_Stop, duration
		}
	}
}

func requestsAbove(elevData Elevator, simRequests map[t.OrderType][]bool) bool {
	for floor := elevData.In_floor + 1; floor < cfg.NumFloors; floor++ {
		if anyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func requestsBelow(elevData Elevator, simRequests map[t.OrderType][]bool) bool {
	for floor := elevData.In_floor - 1; floor >= 0; floor-- {
		if anyRequestsAtFloor(floor, simRequests) {
			return true
		}
	}
	return false
}

func anyRequestsAtFloor(floor int, simRequests map[t.OrderType][]bool) bool {
	return simRequests[t.HallDown][floor] || simRequests[t.GetMyCab(cfg.MyID)][floor] || simRequests[t.HallUp][floor]
}
