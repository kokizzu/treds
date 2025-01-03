package commands

import (
	"fmt"

	"treds/resp"
	"treds/store"
)

const DeletePrefixCommand = "DELPREFIX"

func RegisterDeletePrefixCommand(r CommandRegistry) {
	r.Add(&CommandRegistration{
		Name:     DeletePrefixCommand,
		Validate: validateDeletePrefix(),
		Execute:  executeDeletePrefix(),
		IsWrite:  true,
	})
}

func validateDeletePrefix() ValidationHook {
	return func(args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expected 1 argument, got %d", len(args))
		}
		return nil
	}
}

func executeDeletePrefix() ExecutionHook {
	return func(args []string, store store.Store) string {
		numDel, err := store.DeletePrefix(args[0])
		if err != nil {
			return resp.EncodeError(err.Error())
		}
		return resp.EncodeInteger(numDel)
	}
}
