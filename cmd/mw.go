package main

import (
	"fmt"
	
	"leastauthority.com/mwng/order"
	"leastauthority.com/mwng/key"
)

func main() {
	orderMachine := order.InitMachine(order.StateS0NoPake)
	sortedKeyMachine := key.InitMachine(key.StateS0KnowNothing)

	//dotGraph := orderMachine.ToGraph()
	//fmt.Printf("%s\n", dotGraph)
	orderMachine.Fire(order.TriggerGotNonPake, "side", "version", []byte{0,1,2,3})
	orderMachine.Fire(order.TriggerGotNonPake, "side", "add", []byte{0,1,2,3})
	orderMachine.Fire(order.TriggerGotPake, "side1", "pake", []byte{1,2,3,4})

	// fmt.Printf("%s\n", sortedKeyMachine.ToGraph())
	sortedKeyMachine.Fire(key.TriggerGotCode, "4-purple-sausages", "appID")
}
