package commands

func RegisterCommands(r CommandRegistry) {
	RegisterPINGCommand(r)
	RegisterGetCommand(r)
	RegisterSetCommand(r)
	RegisterMSetCommand(r)
	RegisterDeleteCommand(r)
	RegisterScanKVSCommand(r)
	RegisterScanKeysCommand(r)
	RegisterDeletePrefixCommand(r)
	RegisterKeysCommand(r)
	RegisterKVSCommand(r)
	RegisterMGetCommand(r)
	RegisterDBSizeCommand(r)
	RegisterZAddCommand(r)
	RegisterZRangeLexCommand(r)
	RegisterZRangeLexKeysCommand(r)
	RegisterZRangeScoreCommand(r)
	RegisterZRangeScoreKVSCommand(r)
	RegisterZRemCommand(r)
	RegisterZScoreCommand(r)
	RegisterZCardCommand(r)
	RegisterZRevRangeScoreCommand(r)
	RegisterZRevRangeScoreKVSCommand(r)
	RegisterZRevRangeLexKeysCommand(r)
	RegisterZRevRangeLexKVSCommand(r)
	RegisterFlushAllCommand(r)
	RegisterLPushCommand(r)
	RegisterRPushCommand(r)
	RegisterLPopCommand(r)
	RegisterRPopCommand(r)
	RegisterLRemCommand(r)
	RegisterLSetCommand(r)
	RegisterLRangeCommand(r)
	RegisterLLenCommand(r)
	RegisterLIndexCommand(r)
	RegisterSAddCommand(r)
	RegisterSRemCommand(r)
	RegisterSMembersCommand(r)
	RegisterSIsMemberCommand(r)
	RegisterSCardCommand(r)
	RegisterSUnionCommand(r)
	RegisterSInterCommand(r)
	RegisterSDiffCommand(r)
	RegisterHSetCommand(r)
	RegisterHGetCommand(r)
	RegisterHGetAllCommand(r)
	RegisterHLenCommand(r)
	RegisterHDelCommand(r)
	RegisterHExistsCommand(r)
	RegisterHKeysCommand(r)
	RegisterHValsCommand(r)
	RegisterExpireCommand(r)
	RegisterTtlCommand(r)
	RegisterLongestPrefixCommand(r)
	RegisterKeysHCommand(r)
	RegisterKeysLCommand(r)
	RegisterKeysSCommand(r)
	RegisterKeysZCommand(r)
	RegisterDCreateCollection(r)
	RegisterDInsertCommand(r)
	RegisterDQueryCommand(r)
	RegisterDExplainCommand(r)
	RegisterDDropCollection(r)
	RegisterVCreate(r)
	RegisterVInsert(r)
	RegisterVSearch(r)
	RegisterVDelete(r)
}
