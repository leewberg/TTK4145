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

	txChan := make(chan t.WorldView, 16)
	rxChan := make(chan t.WorldView, 64)

	go bcast.Transmitter(cfg.BroadcastPort, txChan)
	go bcast.Receiver(cfg.BroadcastPort, rxChan)

	go runSender(netID, txChan)
	go runReceiver(netID, rxChan)
}

func runSender(netID string, txChan chan<- t.WorldView) {
	ticker := time.NewTicker(cfg.BroadcastPeriod * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case txChan <- buildWorldSnapshot(netID):
		default:
			fmt.Println("Warn: Network outbox full, dropping packet")
		}

	}
}

func runReceiver(netID string, rxChan <-chan t.WorldView) {
	for msg := range rxChan {

		// Ignore our own broadcast
		if msg.Sender == netID {
			continue
		}

		mergeIncomingWorld(msg)

	}
}

func buildWorldSnapshot(sender string) t.WorldView {
	return t.WorldView{
		Sender:   sender,
		PeerFail: snapshotPeerFailTime(),
		PeerSeen: snapshotPeerSeenTime(),
		Orders:   snapshotAllOrders(),
	}
}

func snapshotPeerFailTime() []int64 {
	failTime := make([]int64, cfg.NumElevators)

	for elevatorID := range cfg.NumElevators {
		failTime[elevatorID] = db.LastMiss(elevatorID)
	}
	return failTime
}

func snapshotPeerSeenTime() []int64 {
	seenTime := make([]int64, cfg.NumElevators)
	for elevatorID := range cfg.NumElevators {
		seenTime[elevatorID] = db.LastSeen(elevatorID)
	}
	return seenTime
}

func snapshotAllOrders() [][]t.OrderData {
	typeCount := cfg.NumElevators + 2 // down, up and one cab per elevator

	out := make([][]t.OrderData, typeCount)

	for ot := range typeCount {
		out[ot] = make([]t.OrderData, cfg.NumFloors)
		for floor := range cfg.NumFloors {
			out[ot][floor] = db.GetOrder(t.OrderType(ot), floor)
		}
	}
	return out
}

func mergeIncomingWorld(in t.WorldView) {
	// Merge peer liveness
	for elevatorID := range in.PeerSeen {
		db.MergePeerSnapshot(elevatorID, in.PeerSeen[elevatorID], in.PeerFail[elevatorID])
	}

	// Merge orders
	typeCount := cfg.NumElevators + 2
	if len(in.Orders) < typeCount {
		return
	}
	for ot := range typeCount {
		if len(in.Orders[ot]) < cfg.NumFloors {
			continue
		}
		for floor := range cfg.NumFloors {
			db.MergeIncomingOrder(t.OrderType(ot), floor, in.Orders[ot][floor])
		}
	}
}
