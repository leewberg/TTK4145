package main

import (
	"fmt"
	assigner "heislabb/source/assignment"
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	elevio "heislabb/source/elevio"
	network "heislabb/source/network"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	if os.Getenv("APP_MODE") == "worker" {
		elevatorMain()
	} else {
		runSupervisor()
	}
}

func runSupervisor() {
	// restarts elevator uppon crash

	signal.Ignore(syscall.SIGTERM)
	for {

		fmt.Println("[Supervisor] Starting application...")

		// spin up the worker
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Env = append(os.Environ(), "APP_MODE=worker")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		// blocking until worker terminates
		err := cmd.Run()

		fmt.Printf("[Supervisor] App crashed or stopped: %v\n", err)
		fmt.Println("[Supervisor] Restarting in 1 second...")
		time.Sleep(1 * time.Second)
	}
}

func elevatorMain() {
	if len(os.Args) < 2 {
		panic("Need one argument, specifying the ID of this elevator")
	}
	var err error
	cfg.MyID, err = strconv.Atoi(os.Args[1])
	if err != nil || cfg.MyID < 0 || cfg.MyID >= cfg.NumElevators {
		panic("ID needs to be an integer between 0 and NUM_ELEVATORS-1")
	}

	elevio.Init("localhost:"+strconv.Itoa(15657+cfg.MyID), cfg.NumFloors)

	elevio.Clear_all_lights()
	db.InitOrders()
	db.InitPeers()
	elevio.LocalElevator.Init()

	time.Sleep(100 * time.Millisecond)
	network.StartNetwork(cfg.MyID)
	go assigner.AssignerRoutine()
	go elevio.Light_routine(cfg.MyID)
	go elevio.ButtonRoutine(&elevio.LocalElevator)
	go elevio.LocalElevator.Elev_routine()

	select {} // blocking without using CPU

}
