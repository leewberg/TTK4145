package elevio

import "fmt"

func printError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
