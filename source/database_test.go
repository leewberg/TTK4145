// Testene er chatta
package elevio

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestServiceAllRequests verifies that every requested order is serviced (cleared)
// only when a simulated elevator calls clearOrder. The test fails if any order
// becomes ORDER_CLEAR without the test's explicit clearOrder call (simulating
// an elevator having serviced it).
//
// This is an integration-style, concurrency-heavy test that exercises assignment,
// re-assignment and concurrent access to the shared databases.
func TestServiceAllRequests(t *testing.T) {
	ensureGlobals(t)

	nFloors := NUM_FLOORS
	if nFloors == 0 {
		t.Fatal("NUM_FLOORS is zero")
	}

	initElevatorData()
	initFunctionalTimes()
	initOrderData()

	// Prepare a set of requested orders on several floors (edge floors included).
	requested := make([]struct {
		ot    OrderType
		floor int
	}, 0)

	// pick a few floors including 0 and top to hit edge cases
	maxToRequest := 6
	for f := 0; f < nFloors && len(requested) < maxToRequest; f++ {
		var ot OrderType
		// alternate hall up/down to exercise direction logic
		if f%2 == 0 {
			ot = HALL_UP
		} else {
			ot = HALL_DOWN
		}
		requested = append(requested, struct {
			ot    OrderType
			floor int
		}{ot: ot, floor: f})
	}

	if len(requested) == 0 {
		t.Skip("no floors to request")
	}

	// Mark orders as requested
	for _, r := range requested {
		requestOrder(r.ot, r.floor)
	}

	// Track clears we perform: key = fmt.Sprintf("%d:%d", ot, floor)
	clearedByUs := make(map[string]bool)
	var cbm sync.Mutex

	// Helper to mark a clear done by our simulated elevator
	markCleared := func(ot OrderType, floor int) {
		cbm.Lock()
		clearedByUs[fmt.Sprintf("%d:%d", ot, floor)] = true
		cbm.Unlock()
	}

	// Monitor that fails if an order becomes CLEAR without our explicit clear
	errCh := make(chan string, 1)
	stopMonitor := make(chan struct{})
	var monitorWG sync.WaitGroup
	monitorWG.Add(1)
	go func() {
		defer monitorWG.Done()
		ticker := time.NewTicker(2 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopMonitor:
				return
			case <-ticker.C:
				for _, r := range requested {
					od := readOrderData(r.ot, r.floor)
					if stateFromVersionNr(od.version_nr) == ORDER_CLEAR {
						key := fmt.Sprintf("%d:%d", r.ot, r.floor)
						cbm.Lock()
						cleared := clearedByUs[key]
						cbm.Unlock()
						if !cleared {
							select {
							case errCh <- fmt.Sprintf("order cleared without calling clearOrder: ot=%v floor=%d", r.ot, r.floor):
							default:
							}
							return
						}
					}
				}
			}
		}
	}()

	// Background assigner: repeatedly run assignOrders to simulate the dispatcher
	stopAssigner := make(chan struct{})
	var assignerWG sync.WaitGroup
	assignerWG.Add(1)
	go func() {
		defer assignerWG.Done()
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopAssigner:
				return
			case <-ticker.C:
				assignOrders()
			}
		}
	}()

	// Start elevator worker goroutines that look for orders assigned to them and clear them.
	var workersWG sync.WaitGroup
	numElev := NUM_ELEVATORS
	if numElev <= 0 {
		numElev = 1
	}
	for elevID := 0; elevID < numElev; elevID++ {
		workersWG.Add(1)
		go func(id int) {
			defer workersWG.Done()
			// run for a limited time or until all requests cleared
			deadline := time.Now().Add(1 * time.Second)
			for time.Now().Before(deadline) {
				allCleared := true
				for _, r := range requested {
					od := readOrderData(r.ot, r.floor)
					st := stateFromVersionNr(od.version_nr)
					if st == ORDER_REQUESTED || st == ORDER_CONFIRMED {
						allCleared = false
					}
					// if this worker is assigned, clear it (simulate servicing)
					if st == ORDER_CONFIRMED && od.assigned_to == id {
						// simulate travel time
						time.Sleep(2 * time.Millisecond)
						clearOrder(r.ot, r.floor)
						markCleared(r.ot, r.floor)
					}
				}
				if allCleared {
					return
				}
				time.Sleep(3 * time.Millisecond)
			}
		}(elevID)
	}

	// Also run a background "external actor" that occasionally merges or touches data
	stopExternal := make(chan struct{})
	var externalWG sync.WaitGroup
	externalWG.Add(1)
	go func() {
		defer externalWG.Done()
		i := 0
		for {
			select {
			case <-stopExternal:
				return
			default:
			}
			// occasionally touch functional times and merge harmless data
			mergeElevFunctionalData(0, time.Now().UnixMilli())
			if i%7 == 0 {
				// merge a non-conflicting higher version requested (should not clear)
				for _, r := range requested {
					mergeOrder(r.ot, r.floor, OrderData{version_nr: readOrderData(r.ot, r.floor).version_nr + 1, assigned_to: -1, assigned_cost: INF})
					break
				}
			}
			i++
			time.Sleep(4 * time.Millisecond)
		}
	}()

	// Wait until all requested orders are cleared by us or until timeout / error.
	timeout := time.After(2 * time.Second)
waitLoop:
	for {
		select {
		case e := <-errCh:
			// unexpected clear detected
			close(stopAssigner)
			close(stopExternal)
			close(stopMonitor)
			assignerWG.Wait()
			externalWG.Wait()
			monitorWG.Wait()
			workersWG.Wait()
			t.Fatalf("unexpected clear: %s", e)
		case <-timeout:
			close(stopAssigner)
			close(stopExternal)
			close(stopMonitor)
			assignerWG.Wait()
			externalWG.Wait()
			monitorWG.Wait()
			workersWG.Wait()
			t.Fatal("timeout: not all requests were serviced within time limit")
		default:
			// check if all requested have been cleared by us
			allCleared := true
			for _, r := range requested {
				key := fmt.Sprintf("%d:%d", r.ot, r.floor)
				cbm.Lock()
				cleared := clearedByUs[key]
				cbm.Unlock()
				if !cleared {
					allCleared = false
					break
				}
			}
			if allCleared {
				break waitLoop
			}
			time.Sleep(5 * time.Millisecond)
		}
	}

	// stop background goroutines gracefully
	close(stopAssigner)
	close(stopExternal)
	close(stopMonitor)
	assignerWG.Wait()
	externalWG.Wait()
	monitorWG.Wait()
	workersWG.Wait()

	// Final verification: every requested order should be ORDER_CLEAR and be in clearedByUs.
	for _, r := range requested {
		od := readOrderData(r.ot, r.floor)
		if stateFromVersionNr(od.version_nr) != ORDER_CLEAR {
			t.Fatalf("order not cleared at end: ot=%v floor=%d state=%v", r.ot, r.floor, stateFromVersionNr(od.version_nr))
		}
		key := fmt.Sprintf("%d:%d", r.ot, r.floor)
		cbm.Lock()
		cleared := clearedByUs[key]
		cbm.Unlock()
		if !cleared {
			t.Fatalf("order cleared but not by our simulated elevators: ot=%v floor=%d", r.ot, r.floor)
		}
	}
}
