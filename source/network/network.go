package network

import (
	"fmt"
	assigner "heislabb/source/assignment"
	config "heislabb/source/config"
	db "heislabb/source/database"
	"heislabb/source/network/bcast"
	types "heislabb/source/types"
	"math/rand/v2"
	"time"
)

func StartNetwork(myID int) {

	netID := fmt.Sprintf("elev-%d", myID)

	outbox := make(chan types.WorldView, 16)
	inbox := make(chan types.WorldView, 64)
	go bcast.Transmitter(config.BCAST_PORT, outbox)
	go bcast.Receiver(config.BCAST_PORT, inbox)

	go func() { // sender
		t := time.NewTicker(config.BROADCAST_PERIOD * time.Millisecond)
		defer t.Stop()

		for {
			<-t.C
			assigner.AssignOrders()
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
			db.RecivedMsg()
			// fmt.Println("Got msg at time", time.Now())
		}
	}()
}

func getWorldSnapshot(sender string) types.WorldView {
	w := types.WorldView{
		Sender:      sender,
		LastFailed:  snapshotFailedTime(),
		ProofOfWork: snapshotProofOfWork(),
		Orders:      snapshotOrdersFlat(),
	}
	return w
}

func snapshotFailedTime() []int64 {
	out := make([]int64, config.NUM_ELEVATORS)
	for i := 0; i < config.NUM_ELEVATORS; i++ {
		out[i] = db.GetLastFailedTime(i)
	}
	return out
}

func snapshotProofOfWork() []int64 {
	out := make([]int64, config.NUM_ELEVATORS)
	for i := 0; i < config.NUM_ELEVATORS; i++ {
		out[i] = db.GetLastProofOfWork(i)
	}
	return out
}

func snapshotOrdersFlat() [][]types.OrderSnapshot {
	nTypes := config.NUM_ELEVATORS + 2
	out := make([][]types.OrderSnapshot, nTypes)

	for t := 0; t < nTypes; t++ {
		out[t] = make([]types.OrderSnapshot, config.NUM_FLOORS)
		for f := 0; f < config.NUM_FLOORS; f++ {
			od := db.ReadOrderData(types.OrderType(t), f)
			out[t][f] = types.OrderSnapshot{
				Version:     od.Version,
				Assigned_to: od.AssignedID,
				Cost:        od.AssignedCost,
				Time:        od.AssignedAtTime,
			}
		}
	}
	return out
}

func mergeNetWorld(in types.WorldView) {
	// Merge elevators
	for ID := range in.ProofOfWork {
		db.MergePeersData(ID, in.ProofOfWork[ID], in.LastFailed[ID])
	}

	// Merge orders
	nTypes := config.NUM_ELEVATORS + 2
	if len(in.Orders) < nTypes {
		return
	}
	for t := 0; t < nTypes; t++ {
		if len(in.Orders[t]) < config.NUM_FLOORS {
			continue
		}
		for f := 0; f < config.NUM_FLOORS; f++ {
			order := in.Orders[t][f]
			db.MergeOrder(types.OrderType(t), f, types.OrderData{
				Version:        order.Version,
				AssignedID:     order.Assigned_to,
				AssignedCost:   order.Cost,
				AssignedAtTime: order.Time,
			})
		}
	}
}
