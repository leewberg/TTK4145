package assignment

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
	"time"
)

func AssignerRoutine() {
	ticker := time.NewTicker(cfg.BroadcastPeriod * time.Millisecond)
	defer ticker.Stop()
	for {
		<-ticker.C
		assignCabOrders()
		assignHallOrders()
	}
}

func assignCabOrders() {
	// assigning cab orders is logically redundant, but it is cleaner than having to deal with AssignedID = -1 only for cab orders
	myCab := t.GetMyCab(cfg.MyID)
	for floor := range cfg.NumFloors {
		order := db.GetOrder(myCab, floor)
		if order.AssignedID == -1 {
			db.AssignToMe(myCab, floor, 0)
		}
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
		isAssigned := order.AssignedID != -1
		isBidWindow := now-order.AssignedTime < cfg.BiddingTime
		hasFailed := isAssigned && now-max(db.LastSeen(order.AssignedID), order.AssignedTime) > cfg.OrderTimeout
		myCost := costFunction(dir, floor)

		if !isAssigned || hasFailed {
			if hasFailed {
				db.LogFailure(order.AssignedID)
			}

			db.AssignToMe(dir, floor, myCost)

		} else if isBidWindow {
			if myCost+cfg.BiddingMinRaise < order.Cost {
				db.AssignToMe(dir, floor, myCost)
			}
		}
	}
}
