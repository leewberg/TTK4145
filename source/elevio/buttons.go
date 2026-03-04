package elevio

import (
	"fmt"
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
)

func ButtonRoutine(e *elevator) {
	var drvButtons = make(chan t.ButtonEvent)
	var drv_floors = make(chan int)
	var drv_obstr = make(chan bool)
	go PollButtons(drvButtons)
	go PollFloorSensor(drv_floors)
	go PollObstructionSwitch(drv_obstr)

	for {
		select {
		case a := <-drvButtons: //hall up, down, or ANY cab button is pressed
			fmt.Printf("%+v\n", a)
			if a.Button == t.BT_HallDown || a.Button == t.BT_HallUp {
				db.ActivateOrder(t.OrderType(a.Button), a.Floor)

			} else { // cab order: adjust to which panel we order from
				db.ActivateOrder(t.OrderType(a.Button)+t.OrderType(cfg.MyID), a.Floor)

			}

		case a := <-drv_floors:
			if a != -1 { //update floor for elevator object if in a floor and not between floors
				if e.isBetweenFloors && e.inFloor != a { //moved into a new floor
					db.Heartbeat()
				}

				e.inFloor = a
				SetFloorIndicator(a)
				e.isBetweenFloors = false

			} else {
				e.isBetweenFloors = true
			}
		}
	}
}
