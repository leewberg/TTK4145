package types

type OrderType int

const (
	HallUp OrderType = iota
	HallDown
	CabFirst
)

func GetMyCab(id int) OrderType {
	return OrderType(id + int(CabFirst))
}

const INF = 2147483647 // 32 bit signed integer limit

type OrderData struct {
	Version int `json:"v"`

	// only relevant in confirmed state
	AssignedID   int   `json:"a"`
	Cost         int   `json:"c"`
	AssignedTime int64 `json:"t"`
}

func (od *OrderData) IsActive() bool {
	return od.Version%2 == 1
}

// World view snapshot
type WorldView struct {
	Sender   string        `json:"sender"`
	PeerSeen []int64       `json:"seen"`
	PeerFail []int64       `json:"fail"`
	Orders   [][]OrderData `json:"orders"`
}
