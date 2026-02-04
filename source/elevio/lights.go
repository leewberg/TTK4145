package elevio

import (
	config "heislabb/source/config"
	db "heislabb/source/database"
	types "heislabb/source/types"
	"time"
)

func Clear_all_lights() {
	for i := range config.NUM_FLOORS {
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
		for i := range config.NUM_FLOORS {
			//check hall buttons
			for j := range 2 {
				order_dir := db.ReadOrderData(types.OrderType(j), i)

				if order_dir.GetState() == types.ORDER_CONFIRMED {
					SetButtonLamp(ButtonType(j), i, true)
				} else {
					SetButtonLamp(ButtonType(j), i, false)

				}
			}
			//check cab button
			ourCab := types.GetMyCab(config.MY_ID)
			order_cab := db.ReadOrderData(ourCab, i)
			if order_cab.GetState() == types.ORDER_CONFIRMED {
				SetButtonLamp(BT_Cab, i, true)
			} else {
				SetButtonLamp(BT_Cab, i, false)
			}
		}
		time.Sleep(100 * time.Millisecond)

	}
}
