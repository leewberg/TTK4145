package network

import (
	"fmt"
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	"heislabb/source/network/bcast"
	t "heislabb/source/types"
	"time"
)

func StartNetwork(myID int) {

	netID := fmt.Sprintf("elev-%d", myID)

	outbox := make(chan t.WorldView, 16)
	inbox := make(chan t.WorldView, 64)
	go bcast.Transmitter(cfg.BroadcastPort, outbox)
	go bcast.Receiver(cfg.BroadcastPort, inbox)

	go func() { // sender
		ticker := time.NewTicker(cfg.BroadcastPeriod * time.Millisecond)
		defer ticker.Stop()

		for {
			<-ticker.C

			select {
			case outbox <- getWorldSnapshot(netID):
			default:
				fmt.Println("Warn: Network outbox full, dropping packet")
			}
			// fmt.Println("State of the order", ReadOrderData(HALL_UP, 2))
			// fmt.Println("elev 0 functional", getFunctionalElevators()[MY_ID])
			// fmt.Println("elev 0 last work", getLastProofOfWork(MY_ID))
			// fmt.Println("elev 0 last fail", getLastFailedTime(MY_ID))
		}
	}()

	go func() { // reciver
		for msg := range inbox {
			if msg.Sender == netID {
				continue
			}

			//simulate packet loss
			if time.Now().UnixNano()%2 == 0 {
				continue
			}

			mergeIncomingWorld(msg)
			// fmt.Println("Got msg at time", time.Now())
		}
	}()
}

func getWorldSnapshot(sender string) t.WorldView {
	return t.WorldView{
		Sender:   sender,
		PeerFail: snapshotPeerFail(),
		PeerSeen: snapshotPeerSeen(),
		Orders:   snapshotOrders(),
	}
}

func snapshotPeerFail() []int64 {
	out := make([]int64, cfg.NumElevators)
	for id := range cfg.NumElevators {
		out[id] = db.LastMiss(id)
	}
	return out
}

func snapshotPeerSeen() []int64 {
	out := make([]int64, cfg.NumElevators)
	for id := range cfg.NumElevators {
		out[id] = db.LastSeen(id)
	}
	return out
}

func snapshotOrders() [][]t.OrderData {
	nTypes := cfg.NumElevators + 2 // down, up and one cab per elevator

	out := make([][]t.OrderData, nTypes)

	for ot := range nTypes {
		out[ot] = make([]t.OrderData, cfg.NumFloors)
		for floor := range cfg.NumFloors {
			out[ot][floor] = db.GetOrder(t.OrderType(ot), floor)
		}
	}
	return out
}

func mergeIncomingWorld(in t.WorldView) {
	// Merge peer liveness
	for ID := range in.PeerSeen {
		db.MergePeerSnapshot(ID, in.PeerSeen[ID], in.PeerFail[ID])
	}

	// Merge orders
	nTypes := cfg.NumElevators + 2
	if len(in.Orders) < nTypes {
		return
	}
	for ot := range nTypes {
		if len(in.Orders[ot]) < cfg.NumFloors {
			continue
		}
		for floor := range cfg.NumFloors {
			db.MergeIncomingOrder(t.OrderType(ot), floor, in.Orders[ot][floor])
		}
	}
}
