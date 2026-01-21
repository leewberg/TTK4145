package elevio

import "fmt"

func ButtonRoutine() {
	drv_buttons := make(chan ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go PollButtons(drv_buttons)
	go PollFloorSensor(drv_floors)
	go PollObstructionSwitch(drv_obstr)
	go PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			/*if a == NUM_FLOORS-1 {
				d = MD_Down
			} else if a == 0 {
				d = MD_Up
			}
			SetMotorDirection(d)*/

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			/*if a {
				SetMotorDirection(MD_Stop)
			} else {
				SetMotorDirection(d)
			}*/

		case a := <-drv_stop:
			//set state as stop for applicable elevator
			fmt.Printf("%+v\n", a)
			for f := 0; f < NUM_FLOORS; f++ {
				for b := ButtonType(0); b < 3; b++ {
					SetButtonLamp(b, f, false)
				}
			}
		}
	}

}
