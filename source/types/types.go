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

const INF = 999000

type OrderData struct {
	Version int `json:"v"`

	// only relevant if active
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

type Elev_states int

const (
	ELEV_BOOT Elev_states = iota
	ELEV_IDLE
	ELEV_RUNNING
	ELEV_DOOR_OPEN
)

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down MotorDirection = -1
	MD_Stop MotorDirection = 0
)

type ButtonType int

const (
	BT_HallUp ButtonType = iota
	BT_HallDown
	BT_Cab
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}
