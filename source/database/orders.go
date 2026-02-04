package database

import (
	// "fmt"
	cfg "heislabb/source/config"
	t "heislabb/source/types"
	"sync"
	"time"
)

var (
	orders      map[t.OrderType][]t.OrderData
	ordersMutex sync.RWMutex
)

func InitOrders() {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	orders = make(map[t.OrderType][]t.OrderData)

	for dir := t.HallUp; dir < cfg.NumElevators+2; dir++ {
		orders[dir] = make([]t.OrderData, cfg.NumFloors)
		for floor := range cfg.NumFloors {
			orders[dir][floor] = t.OrderData{Version: 0, AssignedID: -1, Cost: t.INF, AssignedTime: 0}

		}
	}

}

func GetOrder(dir t.OrderType, floor int) t.OrderData {
	ordersMutex.RLock()
	defer ordersMutex.RUnlock()
	return orders[dir][floor]
}

func RequestOrder(dir t.OrderType, floor int) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	order := &orders[dir][floor]

	if order.GetState() == t.Clear {
		order.Version++
	}
}

func AssignOrder(dir t.OrderType, floor int, cost int) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	// don't take hall orders if im dead
	activePeers := ActiveElevators()
	if dir < t.CabFirst && !activePeers[cfg.MyID] {
		return
	}

	order := &orders[dir][floor]

	if order.GetState() == t.Requested {
		order.Version++ // by assigning the order, we confirm it
	}
	if order.GetState() == t.Confirmed {
		order.Cost = cost
		order.AssignedID = cfg.MyID
		order.AssignedTime = time.Now().UnixMilli()

		if dir < t.CabFirst { // taking a hall order prooves im alive
			Heartbeat()
		}
	}
}

func ClearOrder(dir t.OrderType, floor int) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	order := &orders[dir][floor]

	if order.GetState() == t.Confirmed && order.AssignedID == cfg.MyID {
		order.Version++
		Heartbeat()
		if isPartitioned() {
			order.Version = 0
		}
	}
}

// Consensus logic
func MergeOrder(dir t.OrderType, floor int, incoming t.OrderData) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	if !isValid(incoming) {
		return
	}

	current := orders[dir][floor]

	if incoming.Version > current.Version {

		// Stubbornness clause: you should not externally clear an order assigned to this node
		isMyActiveOrder := current.GetState() == t.Confirmed && current.AssignedID == cfg.MyID
		incomingHasClearedIt := incoming.GetState() != t.Confirmed

		if isMyActiveOrder && incomingHasClearedIt {
			// hijack highest priority
			orders[dir][floor].Version = incoming.Version + (2 - incoming.Version%3)

		} else {
			orders[dir][floor] = incoming

		}

	} else if incoming.Version == current.Version && incoming.GetState() == t.Confirmed {

		currentCost := resolveCost(current)
		incomingCost := resolveCost(incoming)

		if currentCost > incomingCost {
			orders[dir][floor] = incoming
		}

	}
}

// Helpers
func isValid(o t.OrderData) bool {
	if o.GetState() == t.Confirmed &&
		o.AssignedID == -1 {
		return false
	}
	return true
}

func resolveCost(o t.OrderData) int {
	cost := o.Cost
	active := ActiveElevators()
	if !active[o.AssignedID] {
		cost += t.INF
	}
	cost += o.AssignedID // use ID for tiebreaks. ensure ID < MIN_RAISE
	return cost
}
