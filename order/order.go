package order

import (
	"context"
	"reflect"
	"fmt"

	"github.com/qmuntal/stateless"
)

const (
	TriggerGotNonPake = "GotNonPake"
	TriggerGotPake    = "GotPake"
)

const (
	StateS0NoPake = "S0NoPake"
	StateS1YesPake = "S1YesPake"
)

type MsgTuple struct {
	side   string
	phase  string
	body   []byte
}

func InitMachine(initialState stateless.State) *stateless.StateMachine {
	queue := []MsgTuple{}
	order := stateless.NewStateMachine(initialState)

	// gotPake takes "side", "phase" and "body"
	order.SetTriggerParameters(TriggerGotPake, reflect.TypeOf("side"), reflect.TypeOf("phase"), reflect.TypeOf([]byte{}))
	order.SetTriggerParameters(TriggerGotNonPake, reflect.TypeOf("side"), reflect.TypeOf("phase"), reflect.TypeOf([]byte{}))

	// define some simpler state transitions
	order.Configure(StateS0NoPake).
		PermitReentry(TriggerGotNonPake).
		Permit(TriggerGotPake, StateS1YesPake)
	order.Configure(StateS1YesPake).
		PermitReentry(TriggerGotNonPake)

	// define transitions with actions
	// S0_StateNoPake ---> got_pake ---> S1_StateYesPake
	order.Configure(StateS1YesPake).
		OnEntryFrom(TriggerGotPake, func(_ context.Context, args ...interface{}) error {
			// now args[0] is side, args[1] is phase and
			// args[2] is body each of them are "untyped"
			// values (equiv of C (void *).  They will
			// have to be cast to the right type defined
			// above in the SetTriggerParameters.
			//
			// notify_key, drain
			// side := args[0].(string)
			// phase := args[1].(string)
			body := args[2].([]byte)

			// XXX when key FSM is complete, it will be
			// key.GotPake(body)
			fmt.Printf("key.GotPake(%v)\n", body)

			// XXX also drain the (side, phase, body) from
			// queue and call receive machine's
			// GotMessage(side, phase, body)
			for _, msg := range queue {
				fmt.Printf("R.GotMessage(%s, %s, %v)\n",
					msg.side, msg.phase, msg.body)
			}

			return nil
		})

	// StateNoPake -> got_non_pake -> StateNoPake
	order.Configure(StateS0NoPake).
		OnEntryFrom(TriggerGotNonPake, func(_ context.Context, args ...interface{}) error {
			side := args[0].(string)
			phase := args[1].(string)
			body := args[2].([]byte)

			fmt.Printf("adding (%s, %s, %v) to the queue\n", side, phase, body)
			queue = append(queue, MsgTuple{side: side, phase: phase, body: body})
			return nil
		})

	// StateYesPake -> got_non_pake -> StateYesPake
	order.Configure(StateS1YesPake).
		OnEntryFrom(TriggerGotNonPake, func(_ context.Context, args ...interface{}) error {
			//side := args[0].(string)
			//phase := args[1].(string)
			//body := args[2].([]byte)

			// XXX drain the (side, phase, body) from
			// queue and call Receive machine's
			// GotMessage(side, phase, body)
			for _, msg := range queue {
				fmt.Printf("R.GotMessage(%s, %s, %v)\n",
					msg.side, msg.phase, msg.body)
			}
			return nil
		})

	return order
}
