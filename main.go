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
	elevio.LocalElevator.Init(elevio.MY_ID, "dummystring")
	elevio.InitOrderData()
	elevio.InitFunctionalTimes()

	time.Sleep(100 * time.Millisecond)
	elevio.StartNetwork(elevio.MY_ID)
	go elevio.ButtonRoutine(&elevio.LocalElevator)
	go elevio.LocalElevator.Elev_routine()
	go elevio.LocalElevator.Hability_routine()
	go elevio.Light_routine(&elevio.LocalElevator)
	for {
	}

	/*var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)*/

	/*for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			//set state as stop for applicable elevator
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}*/
}
