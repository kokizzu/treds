package commands

import (
	"fmt"
	"strconv"

	"treds/store"
)

const ZCardCommand = "ZCARD"

func RegisterZCardCommand(r CommandRegistry) {
	r.Add(&CommandRegistration{
		Name:     ZCardCommand,
		Validate: validateZCard(),
		Execute:  executeZCardCommand(),
	})
}

func validateZCard() ValidationHook {
	return func(args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("expected 1 argument, got %d", len(args))
		}
		return nil
	}
}

func executeZCardCommand() ExecutionHook {
	return func(args []string, store store.Store) string {
		size, err := store.ZCard(args[0])
		if err != nil {
			return err.Error()
		}
		return strconv.Itoa(size)
	}
}
