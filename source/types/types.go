package types

type OrderState int
type OrderType int

const (
	Clear OrderState = iota
	Requested
	Confirmed
)

const (
	HallUp OrderType = iota
	HallDown
	CabFirst
)

const INF = 2147483647 // 32 bit signed integer limit

type OrderData struct {
	Version int `json:"v"` // contains state info

	// only relevant in confirmed state
	AssignedID   int   `json:"a"`
	Cost         int   `json:"c"`
	AssignedTime int64 `json:"t"`
}

// World view snapshot
type WorldView struct {
	Sender   string        `json:"sender"`
	PeerSeen []int64       `json:"seen"`
	PeerFail []int64       `json:"fail"`
	Orders   [][]OrderData `json:"orders"`
}

func GetMyCab(id int) OrderType {
	return OrderType(id + int(CabFirst))
}

func StateFromVersionNr(vnr int) OrderState {
	if vnr%3 == 0 {
		return Clear
	} else if vnr%3 == 1 {
		return Requested
	} else {
		return Confirmed
	}
}

func (od *OrderData) GetState() OrderState {
	if od.Version%3 == 0 {
		return Clear
	} else if od.Version%3 == 1 {
		return Requested
	} else {
		return Confirmed
	}
}
