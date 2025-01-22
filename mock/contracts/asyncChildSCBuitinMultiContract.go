package contracts

import (
	"math/big"

	"github.com/TerraDharitri/drt-go-chain-vm-common/txDataBuilder"
	mock "github.com/TerraDharitri/drt-go-chain-vm-v2/mock/context"
	test "github.com/TerraDharitri/drt-go-chain-vm-v2/testcommon"
	"github.com/stretchr/testify/require"
)

func childFunctionAsyncChildMock(instanceMock *mock.InstanceMock, config interface{}) {
	testConfig := config.(*AsyncBuiltInCallTestConfig)
	instanceMock.AddMockMethod("childFunction", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		t := instance.T
		arguments := host.Runtime().Arguments()

		host.Metering().UseGas(testConfig.GasUsedByChild)

		destination := arguments[0]
		function := string(arguments[1])
		value := big.NewInt(testConfig.TransferFromChildToParent).Bytes()

		callData := txDataBuilder.NewBuilder()
		callData.Func(function)

		err := host.Runtime().ExecuteAsyncCall(destination, callData.ToBytes(), value)
		require.Nil(t, err)

		return instance
	})
}

func callBackAsyncChildMock(instanceMock *mock.InstanceMock, config interface{}) {
	testConfig := config.(*AsyncBuiltInCallTestConfig)
	instanceMock.AddMockMethod("callBack", test.SimpleWasteGasMockMethod(instanceMock, testConfig.GasUsedByCallback))
}
