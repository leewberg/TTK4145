package elevio

import (
	"time"
)

func Clear_all_lights() {
	for i := range NUM_FLOORS {
		//clear hall buttons
		for j := range 2 {
			SetButtonLamp(ButtonType(j), i, false)
		}
		//clear cab button
		SetButtonLamp(BT_Cab, i, false)
	}
}

func Light_routine(elevID int) {
	for {
		for i := range NUM_FLOORS {
			//check hall buttons
			for j := range 2 {
				order_dir := ReadOrderData(OrderType(j), i)

				if stateFromVersionNr(order_dir.version_nr) == ORDER_CONFIRMED {
					SetButtonLamp(ButtonType(j), i, true)
				} else {
					SetButtonLamp(ButtonType(j), i, false)

				}
			}
			//check cab button
			ourCab := OrderType(2 + elevID)
			order_cab := ReadOrderData(ourCab, i)
			if stateFromVersionNr(order_cab.version_nr) == ORDER_CONFIRMED {
				SetButtonLamp(BT_Cab, i, true)
			} else {
				SetButtonLamp(BT_Cab, i, false)
			}
		}
		time.Sleep(100 * time.Millisecond)

	}
}
