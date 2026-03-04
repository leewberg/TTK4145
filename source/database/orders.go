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
		order.Cost = t.INF
		order.AssignedID = cfg.MyID
		order.AssignedTime = time.Now().UnixMilli()
	}
}

func ClaimOrder(dir t.OrderType, floor int, cost int) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	if !IsFunctional(cfg.MyID) {
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
	}
}

// Consensus logic
func MergeIncomingOrder(dir t.OrderType, floor int, incoming t.OrderData) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	current := orders[dir][floor]

	if incoming.Version > current.Version {

		isMyOrder := current.IsActive() && current.AssignedID == cfg.MyID
		incomingHasClearedIt := !incoming.IsActive()

		// Protects in certain network merging cases
		if isMyOrder && incomingHasClearedIt {
			// ignore and reclaim network priority
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
	// AssignedID for tiebreaks guarantees consistency
	if !IsFunctional(o.AssignedID) {
		return t.INF + o.AssignedID
	}
	return o.Cost + o.AssignedID
}

func GetOrderMatrix(ourCab t.OrderType) map[t.OrderType][]bool {
	orderMatrix := make(map[t.OrderType][]bool)
	now := time.Now().UnixMilli()
	for _, orderType := range []t.OrderType{t.HallUp, t.HallDown, ourCab} {
		orderMatrix[orderType] = make([]bool, cfg.NumFloors)
		for floor := range cfg.NumFloors {
			orderData := GetOrder(orderType, floor)
			if orderData.IsActive() && orderData.AssignedID == cfg.MyID && now-orderData.AssignedTime > cfg.BiddingTime {
				orderMatrix[orderType][floor] = true
			}
		}
	}
	return orderMatrix
}
