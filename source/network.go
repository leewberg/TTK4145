
package elevio

import (
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Elevator snapshot
type NetElev struct {
	ID int `json:"id"`
	F  int `json:"f"` // last_floor
	D  int `json:"d"` // direction
	V  int `json:"v"` // data_version
}

// Order snapshot
type NetOrder struct {
	V int `json:"v"` // version_nr
	A int `json:"a"` // assigned_to
	C int `json:"c"` // assigned_cost
}

// Full world view snapshot
type NetWorld struct {
	Sender string `json:"sender"` // which elevetor that sends the message, ex "elev-0"
	Seq    uint64 `json:"seq"`    // sequence number to prevent duplication of messages or out of order messages

	Elev []NetElev `json:"e"`

	Orders [][]NetOrder `json:"o"`
}



func StartNetwork(myID int) {

	const peerPort = 15647 // tall fra nettverksmodulen
	const bcastPort = 16569

	netID := fmt.Sprintf("elev-%d", myID)

	peerUpdateCh := make(chan peers.PeerUpdate, 16)
	peerTxEnable := make(chan bool, 1)

	go peers.Transmitter(peerPort, netID, peerTxEnable)
	go peers.Receiver(peerPort, peerUpdateCh)

	txWorld := make(chan NetWorld, 16)
	rxWorld := make(chan NetWorld, 64)
	go bcast.Transmitter(bcastPort, txWorld)
	go bcast.Receiver(bcastPort, rxWorld)


	// Sequence number to prevent duplication
	var seq uint64
	lastSeqBySender := make(map[string]uint64)
	var lastSeqMu sync.Mutex

	// Maintain "functional elevators" view based on peers module
	// implement with functionalElevatorManager

	
	go func() {
		t := time.NewTicker(200 * time.Millisecond)
		defer t.Stop()

		for {
			select {
			case <-t.C:
			case <-sendNow:
				// send immediately as well as periodically
			}
			distributeOrders() //does this function exist?

			seq++
			txWorld <- buildNetWorld(netID, seq)
		}
	}()

	// merge snapshots
	go func() {
		for msg := range rxWorld {
			if msg.Sender == netID {
				continue
			}

			// prevent duplication
			lastSeqMu.Lock()
			prev := lastSeqBySender[msg.Sender]
			if msg.Seq <= prev {
				lastSeqMu.Unlock()
				continue
			}
			lastSeqBySender[msg.Sender] = msg.Seq
			lastSeqMu.Unlock()

			mergeNetWorld(msg)

			distributeOrders()
			nonBlockingSignal(sendNow)
		}
	}()
}


func buildNetWorld(sender string, seq uint64) NetWorld {
	w := NetWorld{
		Sender: sender,
		Seq:    seq,
		Elev:   snapshotElevators(),
		Orders: snapshotOrdersFlat(),
	}
	return w
}

func snapshotElevators() []NetElev {
	out := make([]NetElev, NUM_ELEVATORS)
	for i := 0; i < NUM_ELEVATORS; i++ {
		d := getElevData(i)
		out[i] = NetElev{
			ID: i,
			F:  d.last_floor,
			D:  int(d.direction),
			V:  d.data_version,
		}
	}
	return out
}

func snapshotOrdersFlat() [][]NetOrder {
	types := NUM_ELEVATORS + 2
	out := make([][]NetOrder, types)

	for t := 0; t < types; t++ {
		out[t] = make([]NetOrder, NUM_FLOORS)
		for f := 0; f < NUM_FLOORS; f++ {
			od := readOrderData(OrderType(t), f)
			out[t][f] = NetOrder{
				V: od.version_nr,
				A: od.assigned_to,
				C: od.assigned_cost,
			}
		}
	}
	return out
}


func mergeNetWorld(in NetWorld) {
	// Merge elevators
	for _, e := range in.Elev {
		if e.ID < 0 || e.ID >= NUM_ELEVATORS {
			continue
		}
		mergeElevatorData(e.ID, ElevatorData{
			last_floor:   e.F,
			direction:   Direction(e.D),
			data_version: e.V,
		})
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
			mergeOrder(OrderType(t), f, OrderData{
				version_nr:    no.V,
				assigned_to:   no.A,
				assigned_cost: no.C,
			})
		}
	}
}

func setFunctionalFromPeers(myID int, pu peers.PeerUpdate, functional []bool) bool {
	newSet := make([]bool, NUM_ELEVATORS)
	newSet[myID] = true

	for _, pid := range pu.Peers {
		if id, ok := parseElevID(pid); ok {
			if id >= 0 && id < NUM_ELEVATORS {
				newSet[id] = true
			}
		}
	}
