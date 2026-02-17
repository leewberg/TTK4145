package assignment

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
	"time"
)

// import "fmt"

func AssignHallOrders() {
	for _, dir := range []t.OrderType{t.HallUp, t.HallDown} {
		for floor := range cfg.NumFloors {
			assignHallOrder(dir, floor)
		}
	}
}

func assignHallOrder(dir t.OrderType, floor int) {
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
