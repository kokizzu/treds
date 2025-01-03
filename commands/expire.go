package commands

import (
	"fmt"
	"strconv"
	"time"

	"treds/resp"
	"treds/store"
)

const ExpireCommand = "EXPIRE"

func RegisterExpireCommand(r CommandRegistry) {
	r.Add(&CommandRegistration{
		Name:     ExpireCommand,
		Validate: validateExpireCommand(),
		Execute:  executeExpireCommand(),
		IsWrite:  true,
	})
}

func validateExpireCommand() ValidationHook {
	return func(args []string) error {
		if len(args) != 2 {
			_, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			return fmt.Errorf("expected 1 argument, got %d", len(args))
		}
		return nil
	}
}

func executeExpireCommand() ExecutionHook {
	return func(args []string, store store.Store) string {
		key := args[0]
		seconds, _ := strconv.Atoi(args[1])
		now := time.Now()
		expiryTime := now.Add(time.Duration(seconds) * time.Second)
		err := store.Expire(key, expiryTime)
		if err != nil {
			return resp.EncodeError(err.Error())
		}
		return resp.EncodeSimpleString("OK")
	}
}
