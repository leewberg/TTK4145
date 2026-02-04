package main

import (
	"fmt"
	config "heislabb/source/config"
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
	config.MY_ID, err = strconv.Atoi(os.Args[1])
	if err != nil || config.MY_ID < 0 || config.MY_ID >= config.NUM_ELEVATORS {
		panic("ID needs to be an integer between 0 and NUM_ELEVATORS-1")
	}

	elevio.Init("localhost:"+strconv.Itoa(15657+config.MY_ID), config.NUM_FLOORS)

	elevio.Clear_all_lights()
	db.InitOrderData()
	db.InitPeers()
	elevio.LocalElevator.Init(config.MY_ID)

	time.Sleep(100 * time.Millisecond)
	network.StartNetwork(config.MY_ID)
	go elevio.Light_routine(config.MY_ID)
	go elevio.ButtonRoutine(&elevio.LocalElevator)
	go elevio.LocalElevator.Elev_routine()

	select {} // blocking without using CPU

}
