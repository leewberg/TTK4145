package elevio

import (
	cfg "heislabb/source/config"
	db "heislabb/source/database"
	t "heislabb/source/types"
	"time"
)

func ClearAllLights() {
	for i := range cfg.NumFloors {
		//clear hall buttons
		for j := range 2 {
			SetButtonLamp(t.ButtonType(j), i, false)
		}
		//clear cab button
		SetButtonLamp(t.BT_Cab, i, false)
	}
}

func LightRoutine(elevID int) {
	for {
		for i := range cfg.NumFloors {
			//check hall buttons
			for j := range 2 {
				orderDir := db.GetOrder(t.OrderType(j), i)
				SetButtonLamp(t.ButtonType(j), i, orderDir.IsActive())

			}
			//check cab button
			ourCab := t.GetMyCab(cfg.MyID)
			orderCab := db.GetOrder(ourCab, i)
			SetButtonLamp(t.BT_Cab, i, orderCab.IsActive())
		}
		time.Sleep(100 * time.Millisecond)

	}
}
