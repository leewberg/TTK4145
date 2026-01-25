package elevio

import (
	"time"
)

func Light_routine(e *Elevator) {
	for {
		for i := range NUM_FLOORS {
			//check hall buttons
			for j := range 2 {
				order_dir := readOrderData(OrderType(j), i)

				if stateFromVersionNr(order_dir.version_nr) == ORDER_CONFIRMED {
					SetButtonLamp(ButtonType(j), i, true)
				} else {
					SetButtonLamp(ButtonType(j), i, false)

				}
			}
			//check cab button
			order_cab := readOrderData(OrderType(2+e.ID), i)
			if stateFromVersionNr(order_cab.version_nr) == ORDER_CONFIRMED {
				SetButtonLamp(BT_Cab, i, true)
			} else {
				SetButtonLamp(BT_Cab, i, false)
			}
		}
		time.Sleep(100 * time.Millisecond)

	}
}
