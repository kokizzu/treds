package commands

import (
	"treds/resp"
	"treds/store"
)

const PING = "PING"

func RegisterPINGCommand(r CommandRegistry) {
	r.Add(&CommandRegistration{
		Name:     PING,
		Validate: validateDBSize(),
		Execute:  executePINGCommand(),
	})
}

func executePINGCommand() ExecutionHook {
	return func(args []string, store store.Store) string {
		return resp.EncodeSimpleString("PONG")
	}
}
