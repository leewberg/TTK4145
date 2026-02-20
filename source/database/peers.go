package database

import (
	cfg "heislabb/source/config"
	"sync"
	"time"
)

// Module to keep track of liveliness of the elevators
var (
	lastSeen    []int64
	lastFailure []int64
	peerMutex   sync.RWMutex
)

func InitPeers() {
	peerMutex.Lock()
	defer peerMutex.Unlock()

	lastSeen = make([]int64, cfg.NumElevators)
	lastFailure = make([]int64, cfg.NumElevators)
}

func Heartbeat() {
	peerMutex.Lock()
	defer peerMutex.Unlock()

	lastSeen[cfg.MyID] = time.Now().UnixMilli()
}

func LogFailure(id int) {
	peerMutex.Lock()
	defer peerMutex.Unlock()

	lastFailure[id] = time.Now().UnixMilli()
}

func LastSeen(id int) int64 {
	peerMutex.RLock()
	defer peerMutex.RUnlock()

	return lastSeen[id]
}

func LastMiss(id int) int64 {
	peerMutex.RLock()
	defer peerMutex.RUnlock()

	return lastFailure[id]
}

func MergePeerSnapshot(id int, remoteLastSeen int64, remoteLastFail int64) {
	peerMutex.Lock()
	defer peerMutex.Unlock()

	lastSeen[id] = max(lastSeen[id], remoteLastSeen)
	lastFailure[id] = max(lastFailure[id], remoteLastFail)
}

func IsFunctional(id int) bool {
	peerMutex.RLock()
	defer peerMutex.RUnlock()

	now := time.Now().UnixMilli()

	// in case all nodes but one are dead, we need NUM_ELEVATORS-1 timeout cycles to ensure the one functional node has a chance to grab the order
	isResponsive := lastFailure[id] < lastSeen[id]
	failuresAreOld := (now - lastFailure[id]) > cfg.NumElevators*cfg.OrderTimeout

	return isResponsive || failuresAreOld
}
