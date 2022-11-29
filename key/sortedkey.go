package key

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
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
	sp    *gospake2.SPAKE2
}

type pakeMsg struct {
        Body string `json:"pake_v1"`
}

// XXX we should also define GotPake() that gets called from Order FSM
func InitMachine(initialState stateless.State) *stateless.StateMachine {
	sortedKey := stateless.NewStateMachine(initialState)
	sk := &SortedKey{}

	sortedKey.Configure(StateS0KnowNothing).
		Permit(TriggerGotCode, StateS1KnowCode)
	sortedKey.Configure(StateS1KnowCode).
		Permit(TriggerGotPakeGood, StateS2KnowKey).
		Permit(TriggerGotPakeBad, StateS3Scared)

	// types for triggers
	sortedKey.SetTriggerParameters(TriggerGotCode, reflect.TypeOf("code"), reflect.TypeOf("appid"))
	sortedKey.SetTriggerParameters(TriggerGotPakeGood, reflect.TypeOf([]byte{}), reflect.TypeOf("side"))

	// define actions on transitions
	sortedKey.Configure(StateS1KnowCode).
		OnEntryFrom(TriggerGotCode,  func(_ context.Context, args ...interface{}) error {
			// build_pake
			// M.add_message(pake)
			code := args[0].(string)
			appID := args[1].(string)

			return sk.buildAndSendPakeMsg(code, appID)
		})

	sortedKey.Configure(StateS2KnowKey).
		OnEntryFrom(TriggerGotPakeGood, func(_ context.Context, args ...interface{}) error {
			// computekey
			// M.add_message(version)
			// B.got_key
			// R.got_key
			msg2 := args[0].([]byte)
			// side := args[1].(string)
			key, err := sk.sp.Finish(msg2)
			if err != nil {
				log.Printf("spake2: %v\n", err)
				return err
			}

			// XXX: call B.got_key(key)
			fmt.Printf("Got key: %v\n", key)
			phase := "version"
			// dataKey := derivePhaseKey(key, side, phase)

			// XXX use dataKey to encrypt the versions and
			// send the encrypted message to the other
			// side.
			fmt.Printf("M.AddMessage(%s, \"encryptedData\")\n", phase)
			fmt.Printf("R.GotKey(%s)\n", key)

			return nil
		})

		return sortedKey
}


func (sk *SortedKey) buildAndSendPakeMsg(code string, appID string) error {
	// self._sp = SPAKE2_Symmetric(
        //         to_bytes(code), idSymmetric=to_bytes(self._appid))
	// msg1 = self._sp.start()
        // body = dict_to_bytes({"pake_v1": bytes_to_hexstr(msg1)})
        // self._M.add_message("pake", body)
	pw := gospake2.NewPassword(code)
	spake2 := gospake2.SPAKE2Symmetric(pw, gospake2.NewIdentityS(appID))

	sk.sp = &spake2

	body := spake2.Start()
	pakeMsg := pakeMsg{ Body: hex.EncodeToString(body) }
	jsonMsg, err := json.Marshal(pakeMsg)
	if err != nil {
		return err
	}

	// XXX call mailbox.AddMessage()
	fmt.Printf("jsonMsg: %s\n", jsonMsg)
	fmt.Printf("mailbox.AddMessage(\"pake\", %s)\n", hex.EncodeToString(jsonMsg))

	return nil
}
