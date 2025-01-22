package tests

import (
	"testing"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	logger "github.com/TerraDharitri/drt-go-chain-logger"
	vmcommon "github.com/TerraDharitri/drt-go-chain-vm-common"
	"github.com/TerraDharitri/drt-go-chain-vm-common/builtInFunctions"
	"github.com/TerraDharitri/drt-go-chain-vm-v2/config"
	"github.com/TerraDharitri/drt-go-chain-vm-v2/ipc/common"
	"github.com/TerraDharitri/drt-go-chain-vm-v2/ipc/nodepart"
	"github.com/TerraDharitri/drt-go-chain-vm-v2/mock"
	contextmock "github.com/TerraDharitri/drt-go-chain-vm-v2/mock/context"
	worldmock "github.com/TerraDharitri/drt-go-chain-vm-v2/mock/world"
	"github.com/TerraDharitri/drt-go-chain-vm-v2/vmhost"
	"github.com/TerraDharitri/drt-go-chain-vm-v2/vmhost/hostCore"
	"github.com/stretchr/testify/require"
)

var drtVirtualMachine = []byte{5, 0}

func TestVMDriver_DiagnoseWait(t *testing.T) {
	t.Skip("cannot unmarshal BuiltInFuncContainer")

	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(t, blockchain)

	err := driver.DiagnoseWait(100)
	require.Nil(t, err)
}

func TestVMDriver_DiagnoseWaitWithTimeout(t *testing.T) {
	t.Skip("cannot unmarshal BuiltInFuncContainer")

	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(t, blockchain)

	err := driver.DiagnoseWait(5000)
	require.True(t, common.IsCriticalError(err))
	require.Contains(t, err.Error(), "timeout")
	require.True(t, driver.IsClosed())
}

func TestVMDriver_RestartsIfStopped(t *testing.T) {
	t.Skip("cannot unmarshal BuiltInFuncContainer")

	logger.ToggleLoggerName(true)
	_ = logger.SetLogLevel("*:TRACE")

	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(t, blockchain)

	blockchain.GetUserAccountCalled = func(address []byte) (vmcommon.UserAccountHandler, error) {
		return &worldmock.Account{Code: bytecodeCounter}, nil
	}

	vmOutput, err := driver.RunSmartContractCreate(createDeployInput(bytecodeCounter))
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	vmOutput, err = driver.RunSmartContractCall(createCallInput("increment"))
	require.Nil(t, err)
	require.NotNil(t, vmOutput)

	require.False(t, driver.IsClosed())
	_ = driver.Close()
	require.True(t, driver.IsClosed())

	// Per this request, VM is restarted
	vmOutput, err = driver.RunSmartContractCreate(createDeployInput(bytecodeCounter))
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	require.False(t, driver.IsClosed())
}

func BenchmarkVMDriver_RestartsIfStopped(b *testing.B) {
	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(b, blockchain)

	for i := 0; i < b.N; i++ {
		_ = driver.Close()
		require.True(b, driver.IsClosed())
		_ = driver.RestartVMIfNecessary()
		require.False(b, driver.IsClosed())
	}
}

func BenchmarkVMDriver_RestartVMIfNecessary(b *testing.B) {
	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(b, blockchain)

	for i := 0; i < b.N; i++ {
		_ = driver.RestartVMIfNecessary()
	}
}

func TestVMDriver_GetVersion(t *testing.T) {
	t.Skip("cannot unmarshal BuiltInFuncContainer")
	// This test requires `make vm` before running, or must be run directly
	// with `make test`
	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(t, blockchain)
	version := driver.GetVersion()
	require.NotZero(t, len(version))
	require.NotEqual(t, "undefined", version)
}

func newDriver(tb testing.TB, blockchain *contextmock.BlockchainHookStub) *nodepart.VMDriver {
	driver, err := nodepart.NewVMDriver(
		blockchain,
		common.VMArguments{
			VMHostParameters: vmhost.VMHostParameters{
				VMType:               drtVirtualMachine,
				BlockGasLimit:        uint64(10000000),
				GasSchedule:          config.MakeGasMapForTests(),
				ProtectedKeyPrefix:   []byte("N" + "U" + "M" + "B" + "A" + "T"),
				BuiltInFuncContainer: builtInFunctions.NewBuiltInFunctionContainer(),
				EnableEpochsHandler: &mock.EnableEpochsHandlerStub{
					IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
						return flag == hostCore.SCDeployFlag || flag == hostCore.AheadOfTimeGasUsageFlag || flag == hostCore.RepairCallbackFlag || flag == hostCore.BuiltInFunctionsFlag
					},
				},
			},
		},
		nodepart.Config{MaxLoopTime: 1000},
	)
	require.Nil(tb, err)
	require.NotNil(tb, driver)
	require.False(tb, driver.IsClosed())
	return driver
}
