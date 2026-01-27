package main

import (
	// "fmt"
	elevio "heislabb/source"
	"os"
	"strconv"
	"time"
)

func main() {
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
