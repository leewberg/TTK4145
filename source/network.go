package elevio

import (
	"fmt"
	"heislabb/source/network/bcast"
	"time"
)

// Order snapshot
type NetOrder struct {
	V int   `json:"v"` // version_nr
	A int   `json:"a"` // assigned_to
	C int   `json:"c"` // assigned_cost
	T int64 `json:"t"` // assigned_at_time
}

// Full world view snapshot
type WorldView struct {
	Sender string `json:"sender"` // which elevetor that sends the message, ex "elev-0"

	FunctionalElevs []int64 `json:"functionals"`

	Orders [][]NetOrder `json:"o"`
}

func StartNetwork(myID int) {

	const bcastPort = 16569 // tall fra nettverksmodulen

	netID := fmt.Sprintf("elev-%d", myID)

	txWorld := make(chan WorldView, 16)
	rxWorld := make(chan WorldView, 64)
	go bcast.Transmitter(bcastPort, txWorld)
	go bcast.Receiver(bcastPort, rxWorld)

	// implement with functionalElevatorManager

	go func() { // sender
		t := time.NewTicker(BROADCAST_PERIOD * time.Millisecond)
		defer t.Stop()

		for {
			<-t.C
			assignOrders()
			txWorld <- buildNetWorld(netID)
			// fmt.Println("State of the order", ReadOrderData(HALL_UP, 0))
		}
	}()

	// merge snapshots
	go func() {
		for msg := range rxWorld {
			if msg.Sender == netID {
				continue
			}

			mergeNetWorld(msg)
			recivedMsg()
		}
	}()
}

func buildNetWorld(sender string) WorldView {
	w := WorldView{
		Sender:          sender,
		FunctionalElevs: snapshotFuncElevators(),
		Orders:          snapshotOrdersFlat(),
	}
	return w
}

func snapshotFuncElevators() []int64 {
	out := make([]int64, NUM_ELEVATORS)
	for i := 0; i < NUM_ELEVATORS; i++ {
		out[i] = readElevatorFunctional(i)
	}
	return out
}

func snapshotOrdersFlat() [][]NetOrder {
	types := NUM_ELEVATORS + 2
	out := make([][]NetOrder, types)

	for t := 0; t < types; t++ {
		out[t] = make([]NetOrder, NUM_FLOORS)
		for f := 0; f < NUM_FLOORS; f++ {
			od := ReadOrderData(OrderType(t), f)
			out[t][f] = NetOrder{
				V: od.version_nr,
				A: od.assigned_to,
				C: od.assigned_cost,
				T: od.assigned_at_time,
			}
		}
	}
	return out
}

func mergeNetWorld(in WorldView) {
	// Merge elevators
	for ID, e := range in.FunctionalElevs {
		mergeElevFunctionalData(ID, e)
	}

	// Merge orders
	types := NUM_ELEVATORS + 2
	if len(in.Orders) < types {
		return
	}
	for t := 0; t < types; t++ {
		if len(in.Orders[t]) < NUM_FLOORS {
			continue
		}
		for f := 0; f < NUM_FLOORS; f++ {
			no := in.Orders[t][f]
			MergeOrder(OrderType(t), f, OrderData{
				version_nr:       no.V,
				assigned_to:      no.A,
				assigned_cost:    no.C,
				assigned_at_time: no.T,
			})
		}
	}
}
