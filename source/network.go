package elevio

import (
	"fmt"
	"heislabb/source/network/bcast"
	"math/rand/v2"
	"time"
)

// Order snapshot
type OrderSnapshot struct {
	Version     int   `json:"v"` // version_nr
	Assigned_to int   `json:"a"` // assigned_to
	Cost        int   `json:"c"` // assigned_cost
	Time        int64 `json:"t"` // assigned_at_time
}

// Full world view snapshot
type WorldView struct {
	Sender      string            `json:"sender"`
	ProofOfWork []int64           `json:"proofWork"`
	LastFailed  []int64           `json:"lastFailed"`
	Orders      [][]OrderSnapshot `json:"orders"`
}

func StartNetwork(myID int) {

	const bcastPort = 16569 // TODO: put this into config

	netID := fmt.Sprintf("elev-%d", myID)

	outbox := make(chan WorldView, 16)
	inbox := make(chan WorldView, 64)
	go bcast.Transmitter(bcastPort, outbox)
	go bcast.Receiver(bcastPort, inbox)

	go func() { // sender
		t := time.NewTicker(BROADCAST_PERIOD * time.Millisecond)
		defer t.Stop()

		for {
			<-t.C
			assignOrders()
			if rand.IntN(2) == 0 { // simulates packet loss
				continue
			}
			outbox <- getWorldSnapshot(netID)
			// fmt.Println("State of the order", ReadOrderData(HALL_UP, 2))
			// fmt.Println("elev 0 functional", getFunctionalElevators()[MY_ID])
			// fmt.Println("elev 0 last work", getLastProofOfWork(MY_ID))
			// fmt.Println("elev 0 last fail", getLastFailedTime(MY_ID))
		}
	}()

	// merge snapshots
	go func() {
		for msg := range inbox {
			if msg.Sender == netID {
				continue
			}

			mergeNetWorld(msg)
			recivedMsg()
			// fmt.Println("Got msg at time", time.Now())
		}
	}()
}

func getWorldSnapshot(sender string) WorldView {
	w := WorldView{
		Sender:      sender,
		LastFailed:  snapshotFailedTime(),
		ProofOfWork: snapshotProofOfWork(),
		Orders:      snapshotOrdersFlat(),
	}
	return w
}

func snapshotFailedTime() []int64 {
	out := make([]int64, NUM_ELEVATORS)
	for i := 0; i < NUM_ELEVATORS; i++ {
		out[i] = getLastFailedTime(i)
	}
	return out
}

func snapshotProofOfWork() []int64 {
	out := make([]int64, NUM_ELEVATORS)
	for i := 0; i < NUM_ELEVATORS; i++ {
		out[i] = getLastProofOfWork(i)
	}
	return out
}

func snapshotOrdersFlat() [][]OrderSnapshot {
	types := NUM_ELEVATORS + 2
	out := make([][]OrderSnapshot, types)

	for t := 0; t < types; t++ {
		out[t] = make([]OrderSnapshot, NUM_FLOORS)
		for f := 0; f < NUM_FLOORS; f++ {
			od := ReadOrderData(OrderType(t), f)
			out[t][f] = OrderSnapshot{
				Version:     od.version_nr,
				Assigned_to: od.assigned_to,
				Cost:        od.assigned_cost,
				Time:        od.assigned_at_time,
			}
		}
	}
	return out
}

func mergeNetWorld(in WorldView) {
	// Merge elevators
	for ID := range in.ProofOfWork {
		mergeElevFunctionalData(ID, in.ProofOfWork[ID], in.LastFailed[ID])
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
			order := in.Orders[t][f]
			MergeOrder(OrderType(t), f, OrderData{
				version_nr:       order.Version,
				assigned_to:      order.Assigned_to,
				assigned_cost:    order.Cost,
				assigned_at_time: order.Time,
			})
		}
	}
}
