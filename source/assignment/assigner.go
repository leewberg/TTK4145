package assignment

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
	"time"
)

// import "fmt"
func AssignOrders() {
	assignCabOrders()
	assignHallOrders()
}

func assignCabOrders() {
	myCab := t.GetMyCab(cfg.MyID)
	for floor := range cfg.NumFloors {
		order := db.GetOrder(myCab, floor)
		if order.GetState() == t.Requested {
			db.AssignOrder(myCab, floor, 0)
		}
	}
}

func assignHallOrders() {
	for _, dir := range []t.OrderType{t.HallUp, t.HallDown} {
		for floor := range cfg.NumFloors {
			processHallOrder(dir, floor)
		}
	}
}

func processHallOrder(dir t.OrderType, floor int) {
	order := db.GetOrder(dir, floor)
	state := order.GetState()
	now := time.Now().UnixMilli()

	isNewRequest := state == t.Requested
	hasFailed := state == t.Confirmed && now-max(db.LastSeen(order.AssignedID), order.AssignedTime) > cfg.OrderTimeout
	isBidWindow := state == t.Confirmed && now-order.AssignedTime < cfg.BiddingTime

	if isNewRequest || hasFailed {
		if hasFailed {
			db.LogFailure(order.AssignedID)
			if order.AssignedID == cfg.MyID {
				return
			}
		}

		myCost := costFunction(dir, floor)
		db.AssignOrder(dir, floor, myCost)

	} else if isBidWindow {
		myCost := costFunction(dir, floor)
		if myCost+cfg.BiddingMinRaise < order.Cost {
			db.AssignOrder(dir, floor, myCost)
		}
	}
}
