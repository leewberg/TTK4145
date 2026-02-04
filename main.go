package main

import (
	"fmt"
	elevio "heislabb/source"
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

	// make sure the supervisor is not terminated
	signal.Ignore(syscall.SIGTERM)
	for {

		fmt.Println("[Supervisor] Starting application...")

		// spin up the worker
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Env = append(os.Environ(), "APP_MODE=worker")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		// blocking until worker is killed
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
	elevio.MY_ID, err = strconv.Atoi(os.Args[1])
	if err != nil || elevio.MY_ID < 0 || elevio.MY_ID >= elevio.NUM_ELEVATORS {
		panic("ID needs to be an integer between 0 and NUM_ELEVATORS-1")
	}

	elevio.Init("localhost:"+strconv.Itoa(15657+elevio.MY_ID), elevio.NUM_FLOORS)

	elevio.Clear_all_lights()
	elevio.InitOrderData()
	elevio.InitFunctionalTimes()
	elevio.LocalElevator.Init(elevio.MY_ID)

	time.Sleep(100 * time.Millisecond)
	elevio.StartNetwork(elevio.MY_ID)
	go elevio.Light_routine(elevio.MY_ID)
	go elevio.ButtonRoutine(&elevio.LocalElevator)
	go elevio.LocalElevator.Elev_routine()
	for {
	}

}
