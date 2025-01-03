package commands

import (
	"math"
	"strconv"

	"treds/resp"
	"treds/store"
)

const ZRANGESCOREKVS = "ZRANGESCOREKVS"

func RegisterZRangeScoreKVSCommand(r CommandRegistry) {
	r.Add(&CommandRegistration{
		Name:     ZRANGESCOREKVS,
		Validate: validateZRangeScore(),
		Execute:  executeZRangeScoreKVS(),
	})
}

func executeZRangeScoreKVS() ExecutionHook {
	return func(args []string, store store.Store) string {
		startIndex := strconv.Itoa(0)
		if len(args) > 4 {
			startIndex = args[3]
		}
		count := strconv.Itoa(math.MaxInt64)
		if len(args) > 5 {
			count = args[4]
		}
		withScore := true
		if len(args) > 5 {
			includeScore, err := strconv.ParseBool(args[5])
			if err != nil {
				return err.Error()
			}
			withScore = includeScore
		}
		v, err := store.ZRangeByScoreKVS(args[0], args[1], args[2], startIndex, count, withScore)
		if err != nil {
			return resp.EncodeError(err.Error())
		}
		return resp.EncodeStringArray(v)
	}
}
