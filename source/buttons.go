package elevio

import "fmt"

var drv_buttons = make(chan ButtonEvent)
var drv_floors = make(chan int)
var drv_obstr = make(chan bool)
var drv_stop = make(chan bool)

func ButtonRoutine(e *Elevator) {

	go PollButtons(drv_buttons)
	go PollFloorSensor(drv_floors)
	go PollObstructionSwitch(drv_obstr)
	go PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons: //hall up, down, or ANY cab button is pressed
			requestOrder(OrderType(a.Button), a.Floor) //add order to data-matrix

			//TODO: add distiguishability for buttons pressed on the different pannels
			//this may need to be an extension of the requestOrder function

			fmt.Printf("%+v\n", a)
			SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a != -1 { //update floor for elevator object if in a floor and not between floors
				e.in_floor = a
			}

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				e.obstacle = true
			} else {
				e.obstacle = false
			}

		case a := <-drv_stop:
			if a {
				e.state = ELEV_STOP
			} else {
				if e.state == ELEV_STOP {
					e.justStopped = true
				}
			}
			fmt.Printf("%+v\n", a)
			/*for f := 0; f < NUM_FLOORS; f++ {
				for b := ButtonType(0); b < 3; b++ {
					SetButtonLamp(b, f, false)
				}
			}*/
		}
	}

}
