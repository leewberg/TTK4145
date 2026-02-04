package elevio

import (
	"fmt"
	"heislabb/source/network/bcast"
	"math/rand/v2"
	"time"
)

func StartNetwork(myID int) {

	netID := fmt.Sprintf("elev-%d", myID)

	outbox := make(chan WorldView, 16)
	inbox := make(chan WorldView, 64)
	go bcast.Transmitter(BCAST_PORT, outbox)
	go bcast.Receiver(BCAST_PORT, inbox)

	go func() { // sender
		t := time.NewTicker(BROADCAST_PERIOD * time.Millisecond)
		defer t.Stop()

		for {
			<-t.C
			assignOrders()
			if rand.IntN(2) == 0 { // simulates packet loss
				continue
			}

			select {

			case outbox <- getWorldSnapshot(netID):
			default:
				fmt.Println("Warn: Network outbox full, dropping package")
			}
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
				Version:     od.Version,
				Assigned_to: od.AssignedID,
				Cost:        od.AssignedCost,
				Time:        od.AssignedAtTime,
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
				Version:        order.Version,
				AssignedID:     order.Assigned_to,
				AssignedCost:   order.Cost,
				AssignedAtTime: order.Time,
			})
		}
	}
}
