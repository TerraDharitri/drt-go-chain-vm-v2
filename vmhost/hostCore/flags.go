package hostCore

import "github.com/TerraDharitri/drt-go-chain-core/core"

const (
	// SCDeployFlag defines the flag that activates the sc deploy
	SCDeployFlag core.EnableEpochFlag = "SCDeployFlag"
	// BuiltInFunctionsFlag defines the flag that activates the builtin functions
	BuiltInFunctionsFlag core.EnableEpochFlag = "BuiltInFunctionsFlag"
	// RepairCallbackFlag defines the flag that activates the repair callback fix
	RepairCallbackFlag core.EnableEpochFlag = "RepairCallbackFlag"
	// AheadOfTimeGasUsageFlag defines the flag that activates the ahead of time gas usage fix
	AheadOfTimeGasUsageFlag core.EnableEpochFlag = "AheadOfTimeGasUsageFlag"
)

// allFlags must have all flags used by drt-go-chain-vm-v2 in the current version
var allFlags = []core.EnableEpochFlag{
	SCDeployFlag,
	BuiltInFunctionsFlag,
	RepairCallbackFlag,
	AheadOfTimeGasUsageFlag,
}
