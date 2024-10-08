package commands

import (
	"treds/store"
)

const LongestPrefixCommand = "LNGPREFIX"

func RegisterLongestPrefixCommand(r CommandRegistry) {
	r.Add(&CommandRegistration{
		Name:     LongestPrefixCommand,
		Validate: validateDeletePrefix(),
		Execute:  executeLongestPrefixCommand(),
	})
}

func executeLongestPrefixCommand() ExecutionHook {
	return func(args []string, store store.Store) string {
		prefix := args[0]
		res, err := store.LongestPrefix(prefix)
		if err != nil {
			return err.Error()
		}
		return res
	}
}
