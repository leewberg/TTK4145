// Testene er chatta
package elevio

import (
	"sync"
	"testing"
	"time"
)

func makeSimRequests(ourCab OrderType) map[OrderType][]bool {
	// create map for HALL_UP, HALL_DOWN and ourCab with slices sized to NUM_FLOORS
	sim := make(map[OrderType][]bool)
	sim[HALL_UP] = make([]bool, int(NUM_FLOORS))
	sim[HALL_DOWN] = make([]bool, int(NUM_FLOORS))
	sim[ourCab] = make([]bool, int(NUM_FLOORS))
	return sim
}

func TestChooseDirection_BasicCases(t *testing.T) {
	ourCab := OrderType(CAB_FIRST + 1)

	t.Run("order below prefers down", func(t *testing.T) {
		sim := makeSimRequests(ourCab)
		// place order below
		if len(sim[HALL_DOWN]) <= 1 {
			t.Fatalf("NUM_FLOORS too small for test")
		}
		sim[HALL_DOWN][1] = true

		elev := ElevatorData{
			last_floor: 2,
			direction:  DIR_UP,
		}

		got := chooseDirection(elev, sim, ourCab)
		if got != DIR_DOWN {
			t.Errorf("chooseDirection() = %v; want DIR_DOWN", got)
		}
	})

	t.Run("order above with current direction up prefers up", func(t *testing.T) {
		sim := makeSimRequests(ourCab)
		// place order above
		if len(sim[HALL_UP]) <= 3 {
			t.Fatalf("NUM_FLOORS too small for test")
		}
		sim[HALL_UP][3] = true

		elev := ElevatorData{
			last_floor: 2,
			direction:  DIR_UP,
		}

		got := chooseDirection(elev, sim, ourCab)
		if got != DIR_UP {
			t.Errorf("chooseDirection() = %v; want DIR_UP", got)
		}
	})

	t.Run("no orders respects current down direction", func(t *testing.T) {
		sim := makeSimRequests(ourCab)

		elev := ElevatorData{
			last_floor: 1,
			direction:  DIR_DOWN,
		}

		got := chooseDirection(elev, sim, ourCab)
		if got != DIR_DOWN {
			t.Errorf("chooseDirection() = %v; want DIR_DOWN", got)
		}
	})
}

func TestChooseDirection_MoreCases(t *testing.T) {
	ourCab := OrderType(CAB_FIRST + 1)

	t.Run("order above flips to up when currently down", func(t *testing.T) {
		sim := makeSimRequests(ourCab)
		if len(sim[HALL_UP]) <= 3 {
			t.Fatalf("NUM_FLOORS too small for test")
		}
		sim[HALL_UP][3] = true

		elev := ElevatorData{
			last_floor: 2,
			direction:  DIR_DOWN,
		}
		got := chooseDirection(elev, sim, ourCab)
		if got != DIR_UP {
			t.Errorf("chooseDirection() = %v; want DIR_UP", got)
		}
	})

	t.Run("both above and below prefers below", func(t *testing.T) {
		sim := makeSimRequests(ourCab)
		if len(sim[HALL_UP]) <= 3 || len(sim[HALL_DOWN]) <= 1 {
			t.Fatalf("NUM_FLOORS too small for test")
		}
		// place one below and one above the elevator
		sim[HALL_DOWN][1] = true
		sim[HALL_UP][3] = true

		elev := ElevatorData{
			last_floor: 2,
			direction:  DIR_UP,
		}

		got := chooseDirection(elev, sim, ourCab)
		if got != DIR_UP {
			t.Errorf("chooseDirection() = %v; want DIR_UP (prefer above when both exist)", got)
		}
	})

	t.Run("top floor with no orders goes down", func(t *testing.T) {
		sim := makeSimRequests(ourCab)
		top := len(sim[HALL_UP]) - 1
		elev := ElevatorData{
			last_floor: top,
			direction:  DIR_UP,
		}

		got := chooseDirection(elev, sim, ourCab)
		if got != DIR_DOWN {
			t.Errorf("chooseDirection() at top floor = %v; want DIR_DOWN", got)
		}
	})
}

func TestElevShouldStop_BasicCases(t *testing.T) {
	ourCab := OrderType(CAB_FIRST + 1)

	t.Run("stop for cab order at current floor", func(t *testing.T) {
		sim := makeSimRequests(ourCab)
		if len(sim[ourCab]) <= 2 {
			t.Fatalf("NUM_FLOORS too small for test")
		}
		sim[ourCab][2] = true

		elev := ElevatorData{
			last_floor: 2,
			direction:  DIR_UP,
		}

		if !elevShouldStop(elev, sim, ourCab) {
			t.Errorf("elevShouldStop() = false; want true for cab order at current floor")
		}
	})

	t.Run("stop for hall order matching travel direction", func(t *testing.T) {
		sim := makeSimRequests(ourCab)
		if len(sim[HALL_UP]) <= 2 {
			t.Fatalf("NUM_FLOORS too small for test")
		}
		sim[HALL_UP][2] = true

		elev := ElevatorData{
			last_floor: 2,
			direction:  DIR_UP,
		}

		if !elevShouldStop(elev, sim, ourCab) {
			t.Errorf("elevShouldStop() = false; want true for hall up order at current floor when direction is up")
		}
	})

	t.Run("do not stop when no requests at current floor", func(t *testing.T) {
		sim := makeSimRequests(ourCab)

		elev := ElevatorData{
			last_floor: 1,
			direction:  DIR_UP,
		}

		if elevShouldStop(elev, sim, ourCab) {
			t.Errorf("elevShouldStop() = true; want false when no requests at current floor")
		}
	})
}

func TestElevShouldStop_MoreCases(t *testing.T) {
	ourCab := OrderType(CAB_FIRST + 1)

	t.Run("no requests -> no stop", func(t *testing.T) {
		sim := makeSimRequests(ourCab)

		elev := ElevatorData{
			last_floor: 1,
			direction:  DIR_UP,
		}

		if elevShouldStop(elev, sim, ourCab) {
			t.Errorf("elevShouldStop() = true; want false when no requests present")
		}
	})

	t.Run("stop for hall down when going down", func(t *testing.T) {
		sim := makeSimRequests(ourCab)
		if len(sim[HALL_DOWN]) <= 2 {
			t.Fatalf("NUM_FLOORS too small for test")
		}
		sim[HALL_DOWN][2] = true

		elev := ElevatorData{
			last_floor: 2,
			direction:  DIR_DOWN,
		}

		if !elevShouldStop(elev, sim, ourCab) {
			t.Errorf("elevShouldStop() = false; want true for HALL_DOWN matching current direction")
		}
	})

	t.Run("stop for cab order at current floor", func(t *testing.T) {
		sim := makeSimRequests(ourCab)
		if len(sim[ourCab]) <= 0 {
			t.Fatalf("NUM_FLOORS too small for test")
		}
		sim[ourCab][0] = true

		elev := ElevatorData{
			last_floor: 0,
			direction:  DIR_UP,
		}

		if !elevShouldStop(elev, sim, ourCab) {
			t.Errorf("elevShouldStop() = false; want true for cab order at current floor")
		}
	})
}

func ensureGlobals(t *testing.T) {
	t.Helper()

	// Initialize orders map with keys we use in tests.
	mutexOD.Lock()
	if allOrdersData == nil {
		allOrdersData = make(map[OrderType][]OrderData)
	}
	// create slices sized to number of floors
	nFloors := NUM_FLOORS
	for ot := HALL_UP; ot < NUM_ELEVATORS+2; ot++ {
		if _, ok := allOrdersData[ot]; !ok || len(allOrdersData[ot]) != nFloors {
			allOrdersData[ot] = make([]OrderData, nFloors)
		}
		for f := 0; f < nFloors; f++ {
			allOrdersData[ot][f] = OrderData{version_nr: 0, assigned_to: -1, assigned_cost: INF}
		}
	}
	mutexOD.Unlock()

	// initialize elevator state and functional times
	InitElevatorData()
	InitFunctionalTimes()

	// mark all elevators functional now
	now := time.Now().UnixMilli()
	for i := range allElevatorsData {
		mergeElevFunctionalData(i, now)
		// give them sensible default state
		mergeElevatorData(i, ElevatorData{last_floor: 0, state: STATE_IDLE, direction: DIR_UP, data_version: 1})
	}
}

// TestAssignOrders_Basic verifies that a requested hall order is assigned to some functional elevator.
func TestAssignOrders_Basic(t *testing.T) {
	ensureGlobals(t)

	// create a requested hall order at floor 1
	if NUM_FLOORS < 2 {
		t.Skip("not enough floors to run test")
	}
	requestFloor := 1

	// set the order to REQUESTED
	mutexOD.Lock()
	allOrdersData[HALL_UP][requestFloor].version_nr = 1 // ORDER_REQUESTED
	allOrdersData[HALL_UP][requestFloor].assigned_to = -1
	allOrdersData[HALL_UP][requestFloor].assigned_cost = INF
	mutexOD.Unlock()

	// ensure elevators have distinct positions so costFunction can differentiate
	for i := range allElevatorsData {
		mergeElevatorData(i, ElevatorData{last_floor: i, state: STATE_IDLE, direction: DIR_UP, data_version: 2})
	}
	setElevData(ElevatorData{last_floor: 2, state: STATE_IDLE, direction: DIR_UP, data_version: 2})

	// run assignment
	assignOrders()

	// verify order was assigned to some functional elevator and is confirmed
	res := readOrderData(HALL_UP, requestFloor)
	if stateFromVersionNr(res.version_nr) != ORDER_CONFIRMED {
		t.Fatalf("order not confirmed after assignOrders; version_nr=%d", res.version_nr)
	}
	funcs := getFunctionalElevators()
	if res.assigned_to < 0 || res.assigned_to >= len(funcs) {
		t.Fatalf("assigned_to out of range: %d", res.assigned_to)
	}
	if !funcs[res.assigned_to] {
		t.Fatalf("order was assigned to a non-functional elevator %d", res.assigned_to)
	}
	if res.assigned_cost >= INF {
		t.Fatalf("assigned_cost not set or INF: %d", res.assigned_cost)
	}
}

// TestAssignOrders_ConcurrentAccess runs assignOrders while another goroutine repeatedly
// performs reads and writes on the order/elevator databases to simulate other programs.
func TestAssignOrders_ConcurrentAccess(t *testing.T) {
	ensureGlobals(t)

	if NUM_FLOORS < 2 {
		t.Skip("not enough floors to run test")
	}

	// create several requested orders on different floors
	mutexOD.Lock()
	for f := 0; f < NUM_FLOORS && f < 3; f++ {
		allOrdersData[HALL_DOWN][f].version_nr = 1 // REQUESTED
		allOrdersData[HALL_DOWN][f].assigned_to = -1
		allOrdersData[HALL_DOWN][f].assigned_cost = INF
	}
	mutexOD.Unlock()

	// start background goroutine that mimics another program accessing the DBs
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		for {
			select {
			case <-stop:
				return
			default:
			}
			// alternate requesting and clearing an order
			f := i % NUM_FLOORS
			requestOrder(HALL_UP, f)
			_ = readOrderData(HALL_UP, f)
			// merge a fake external order occasionally
			if i%5 == 0 {
				mergeOrder(HALL_UP, f, OrderData{version_nr: 2, assigned_to: 0, assigned_cost: 10})
			}
			// touch elevator functional times
			mergeElevFunctionalData(0, time.Now().UnixMilli())
			time.Sleep(2 * time.Millisecond)
			i++
		}
	}()

	// run multiple assignment cycles while background goroutine is active
	for k := 0; k < 10; k++ {
		assignOrders()
		time.Sleep(5 * time.Millisecond)
	}

	// stop background and wait
	close(stop)
	wg.Wait()

	// sanity-check: every confirmed order should be assigned to a functional elevator
	funcs := getFunctionalElevators()
	for _, ot := range []OrderType{HALL_DOWN, HALL_UP} {
		for f := 0; f < NUM_FLOORS; f++ {
			od := readOrderData(ot, f)
			if stateFromVersionNr(od.version_nr) == ORDER_CONFIRMED {
				if od.assigned_to < 0 || od.assigned_to >= len(funcs) {
					t.Fatalf("confirmed order assigned_to out of range: %v (ot=%v,f=%d)", od.assigned_to, ot, f)
				}
				if !funcs[od.assigned_to] {
					t.Fatalf("confirmed order assigned to non-functional elevator %d (ot=%v,f=%d)", od.assigned_to, ot, f)
				}
			}
		}
	}
}
