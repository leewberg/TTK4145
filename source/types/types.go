package types

type OrderState int
type OrderType int

const (
	ORDER_CLEAR OrderState = iota
	ORDER_REQUESTED
	ORDER_CONFIRMED
)

const (
	HALL_UP OrderType = iota
	HALL_DOWN
	CAB_FIRST
)

const INF = 2147483647 // 32 bit signed integer limit

type OrderData struct {
	Version int // contains state info

	// only relevant in confirmed state
	AssignedID     int
	AssignedCost   int
	AssignedAtTime int64
}

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

func GetMyCab(id int) OrderType {
	return OrderType(id + int(CAB_FIRST))
}

func StateFromVersionNr(vnr int) OrderState {
	if vnr%3 == 0 {
		return ORDER_CLEAR
	} else if vnr%3 == 1 {
		return ORDER_REQUESTED
	} else {
		return ORDER_CONFIRMED
	}
}

func (od *OrderData) GetState() OrderState {
	if od.Version%3 == 0 {
		return ORDER_CLEAR
	} else if od.Version%3 == 1 {
		return ORDER_REQUESTED
	} else {
		return ORDER_CONFIRMED
	}
}
