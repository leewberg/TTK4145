package elevio

import (
	"fmt"
	config "heislabb/source/config"
	db "heislabb/source/database"
	types "heislabb/source/types"
	"time"
)

func ButtonRoutine(e *Elevator) {
	var drv_buttons = make(chan ButtonEvent)
	var drv_floors = make(chan int)
	var drv_obstr = make(chan bool)
	go PollButtons(drv_buttons)
	go PollFloorSensor(drv_floors)
	go PollObstructionSwitch(drv_obstr)

	for {
		select {
		case a := <-drv_buttons: //hall up, down, or ANY cab button is pressed
			fmt.Printf("%+v\n", a)
			if a.Button == BT_HallDown || a.Button == BT_HallUp {
				db.RequestOrder(types.OrderType(a.Button), a.Floor)

			} else { // cab order: adjust to which panel we order from
				db.RequestOrder(types.OrderType(a.Button)+types.OrderType(config.MY_ID), a.Floor)

			}

		case a := <-drv_floors:
			if a != -1 { //update floor for elevator object if in a floor and not between floors
				if e.Is_between_floors && e.In_floor != a { //moved into a new floor
					db.WorkProven()
				}

				e.In_floor = a
				SetFloorIndicator(a)
				e.Is_between_floors = false

			} else {
				e.Is_between_floors = true
			}
		}
		if GetObstruction() {
			e.doorOpenTime = time.Now()
		}
	}

}
