package database

import (
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

func ActivateOrder(dir t.OrderType, floor int) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	order := &orders[dir][floor]

	if !order.IsActive() {
		order.Version++
		order.AssignedID = -1
		order.Cost = t.INF
		order.AssignedTime = 0
	}
}

func AssignToMe(dir t.OrderType, floor int, cost int) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	// don't take hall orders if im dead
	activePeers := ActiveElevators()
	if dir < t.CabFirst && !activePeers[cfg.MyID] {
		return
	}

	order := &orders[dir][floor]

	if order.IsActive() {
		order.Cost = cost
		order.AssignedID = cfg.MyID
		order.AssignedTime = time.Now().UnixMilli()
	}
}

func ClearOrder(dir t.OrderType, floor int) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	order := &orders[dir][floor]

	if order.IsActive() && order.AssignedID == cfg.MyID {
		order.Version++
		Heartbeat()
		if isPartitioned() {
			// partitioned networks have lower priority: they can't take hall orders
			order.Version = 0
		}
	}
}

// Consensus logic
func MergeIncomingOrder(dir t.OrderType, floor int, incoming t.OrderData) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	current := orders[dir][floor]

	if incoming.Version > current.Version {

		// You should not externally clear my cab order
		isMyCab := current.IsActive() && dir == t.GetMyCab(cfg.MyID)
		incomingHasClearedIt := !incoming.IsActive()

		if isMyCab && incomingHasClearedIt {
			orders[dir][floor].Version = incoming.Version + 1

		} else {
			orders[dir][floor] = incoming

		}

	} else if incoming.Version == current.Version && incoming.IsActive() {

		currentCost := resolveCost(current)
		incomingCost := resolveCost(incoming)

		if currentCost > incomingCost {
			orders[dir][floor] = incoming
		}

	}
}

func resolveCost(o t.OrderData) int {
	if o.AssignedID == -1 {
		return t.INF
	}

	cost := o.Cost
	active := ActiveElevators()
	if !active[o.AssignedID] {
		cost += t.INF
	}
	cost += o.AssignedID // tiebreak
	return cost
}
