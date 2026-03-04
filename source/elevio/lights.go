package elevio

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
	"time"
)

func Clear_all_lights() {
	for i := range cfg.NumFloors {
		//clear hall buttons
		for j := range 2 {
			SetButtonLamp(t.ButtonType(j), i, false)
		}
		//clear cab button
		SetButtonLamp(t.BT_Cab, i, false)
	}
}

func Light_routine(elevID int) {
	for {
		for i := range cfg.NumFloors {
			//check hall buttons
			for j := range 2 {
				order_dir := db.GetOrder(t.OrderType(j), i)
				SetButtonLamp(t.ButtonType(j), i, order_dir.IsActive())

			}
			//check cab button
			ourCab := t.GetMyCab(cfg.MyID)
			order_cab := db.GetOrder(ourCab, i)
			SetButtonLamp(t.BT_Cab, i, order_cab.IsActive())
		}
		time.Sleep(100 * time.Millisecond)

	}
}
