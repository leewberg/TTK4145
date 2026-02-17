package assignment

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
	"time"
)

func AssignOrders() {
	assignCabOrders()
	assignHallOrders()
}

func assignCabOrders() {
	// this is simpler than having to account for only hall-orders having a non -1 AssignedID
	myCab := t.GetMyCab(cfg.MyID)
	for floor := range cfg.NumFloors {
		order := db.GetOrder(myCab, floor)
		if order.AssignedID == -1 {
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
	now := time.Now().UnixMilli()

	if order.IsActive() {
		isAssigned := order.AssignedID != -1
		isBidWindow := now-order.AssignedTime < cfg.BiddingTime
		hasFailed := isAssigned && now-max(db.LastSeen(order.AssignedID), order.AssignedTime) > cfg.OrderTimeout
		myCost := costFunction(dir, floor)

		if !isAssigned || hasFailed {
			if hasFailed {
				db.LogFailure(order.AssignedID)
			}

			db.AssignOrder(dir, floor, myCost)

		} else if isBidWindow {
			if myCost+cfg.BiddingMinRaise < order.Cost {
				db.AssignOrder(dir, floor, myCost)
			}
		}
	}
}
