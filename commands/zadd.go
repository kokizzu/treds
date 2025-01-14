package commands

import (
	"fmt"

	"treds/resp"
	"treds/store"
)

const ZAddCommand = "ZADD"

func RegisterZAddCommand(r CommandRegistry) {
	r.Add(&CommandRegistration{
		Name:     ZAddCommand,
		Validate: validateZAddCommand(),
		Execute:  executeZAddCommand(),
		IsWrite:  true,
	})
}

func validateZAddCommand() ValidationHook {
	return func(args []string) error {
		if len(args) < 3 {
			return fmt.Errorf("expected 3 or multiple of 3 arguments, got %d", len(args))
		}
		return nil
	}
}

func executeZAddCommand() ExecutionHook {
	return func(args []string, store store.Store) string {
		err := store.ZAdd(args)
		if err != nil {
			return resp.EncodeError(err.Error())
		}
		return resp.EncodeSimpleString("OK")
	}
}
