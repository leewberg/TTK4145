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
			SetButtonLamp(ButtonType(j), i, false)
		}
		//clear cab button
		SetButtonLamp(BT_Cab, i, false)
	}
}

func Light_routine(elevID int) {
	for {
		for i := range cfg.NumFloors {
			//check hall buttons
			for j := range 2 {
				order_dir := db.GetOrder(t.OrderType(j), i)

				if order_dir.GetState() == t.Confirmed {
					SetButtonLamp(ButtonType(j), i, true)
				} else {
					SetButtonLamp(ButtonType(j), i, false)

				}
			}
			//check cab button
			ourCab := t.GetMyCab(cfg.MyID)
			order_cab := db.GetOrder(ourCab, i)
			if order_cab.GetState() == t.Confirmed &&
				time.Now().UnixMilli()-order_cab.AssignedTime > cfg.PartitionTimeout {
				SetButtonLamp(BT_Cab, i, true)
			} else {
				SetButtonLamp(BT_Cab, i, false)
			}
		}
		time.Sleep(100 * time.Millisecond)

	}
}
