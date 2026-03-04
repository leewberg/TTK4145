package elevio

import (
	"fmt"
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
	"time"
)

func ButtonRoutine(e *elevator) {
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
				db.ActivateOrder(t.OrderType(a.Button), a.Floor)

			} else { // cab order: adjust to which panel we order from
				db.ActivateOrder(t.OrderType(a.Button)+t.OrderType(cfg.MyID), a.Floor)

			}

		case a := <-drv_floors:
			if a != -1 { //update floor for elevator object if in a floor and not between floors
				if e.is_between_floors && e.in_floor != a { //moved into a new floor
					db.Heartbeat()
				}

				e.in_floor = a
				SetFloorIndicator(a)
				e.is_between_floors = false

			} else {
				e.is_between_floors = true
			}
		}
		if getObstruction() {
			e.doorOpenTime = time.Now()
		}
	}

}
