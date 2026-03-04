package assignment

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	elevio "heislabb/source/elevio"
	t "heislabb/source/types"
	"time"
)

func AssignerRoutine() {
	ticker := time.NewTicker(cfg.BroadcastPeriod * time.Millisecond)
	defer ticker.Stop()
	for {
		<-ticker.C
		assignHallOrders()
	}
}

func assignHallOrders() {
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
		myCost := elevio.CostFunction(dir, floor)
		isBidWindow := now-order.AssignedTime < cfg.BiddingTime
		hasLowerbid := myCost+cfg.BiddingMinRaise < order.Cost
		hasTimedOut := now-max(db.LastSeen(order.AssignedID), order.AssignedTime) > cfg.OrderTimeout

		if hasTimedOut {
			db.LogFailure(order.AssignedID)
			db.ClaimOrder(dir, floor, myCost)

		} else if isBidWindow && hasLowerbid {
			db.ClaimOrder(dir, floor, myCost)
		}
	}
}
