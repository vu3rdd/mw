package key

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/qmuntal/stateless"
	"salsa.debian.org/vasudev/gospake2"
)

// define triggers (aka events)
const (
	TriggerGotCode     = "GotCode"
	TriggerGotPakeGood = "GotPakeGood"
	TriggerGotPakeBad  = "GotPakeBad"
)

// define states
const (
	StateS0KnowNothing = "S0_KnowNothing"
	StateS1KnowCode    = "S1_KnowCode"
	StateS2KnowKey     = "S2_KnowKey"
	StateS3Scared      = "S3_Scared"
)

type SortedKey struct {
	appid string
	sp    gospake2.SPAKE2
}

type pakeMsg struct {
        Body string `json:"pake_v1"`
}

// XXX we should also define GotPake() that gets called from Order FSM
func InitMachine(initialState stateless.State) *stateless.StateMachine {
	sortedKey := stateless.NewStateMachine(initialState)

	sortedKey.Configure(StateS0KnowNothing).
		Permit(TriggerGotCode, StateS1KnowCode)
	sortedKey.Configure(StateS1KnowCode).
		Permit(TriggerGotPakeGood, StateS2KnowKey).
		Permit(TriggerGotPakeBad, StateS3Scared)

	// types for triggers
	sortedKey.SetTriggerParameters(TriggerGotCode, reflect.TypeOf("code"), reflect.TypeOf("appid"))

	// define actions on transitions
	sortedKey.Configure(StateS1KnowCode).
		OnEntryFrom(TriggerGotCode,  func(_ context.Context, args ...interface{}) error {
			code := args[0].(string)
			appID := args[1].(string)

			return buildAndSendPakeMsg(code, appID)
		})

	return sortedKey
}

func buildAndSendPakeMsg(code string, appID string) error {
	// self._sp = SPAKE2_Symmetric(
        //         to_bytes(code), idSymmetric=to_bytes(self._appid))
	// msg1 = self._sp.start()
        // body = dict_to_bytes({"pake_v1": bytes_to_hexstr(msg1)})
        // self._M.add_message("pake", body)
	pw := gospake2.NewPassword(code)
	spake2 := gospake2.SPAKE2Symmetric(pw, gospake2.NewIdentityS(appID))

	body := spake2.Start()
	pakeMsg := pakeMsg{ Body: hex.EncodeToString(body) }
	jsonMsg, err := json.Marshal(pakeMsg)
	if err != nil {
		return err
	}

	// XXX call mailbox.AddMessage()
	fmt.Printf("jsonMsg: %s\n", jsonMsg)
	fmt.Printf("mailbox.AddMessage(\"pake\", %s)", hex.EncodeToString(jsonMsg))

	return nil
}
