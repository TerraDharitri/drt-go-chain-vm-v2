package scenarioexec

import (
	"fmt"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	logger "github.com/TerraDharitri/drt-go-chain-logger"
	vmi "github.com/TerraDharitri/drt-go-chain-vm-common"
	"github.com/TerraDharitri/drt-go-chain-vm-v2/config"
	"github.com/TerraDharitri/drt-go-chain-vm-v2/mock"
	worldhook "github.com/TerraDharitri/drt-go-chain-vm-v2/mock/world"
	gasSchedules "github.com/TerraDharitri/drt-go-chain-vm-v2/scenarioexec/gasSchedules"
	mc "github.com/TerraDharitri/drt-go-chain-vm-v2/scenarios/controller"
	er "github.com/TerraDharitri/drt-go-chain-vm-v2/scenarios/expression/reconstructor"
	fr "github.com/TerraDharitri/drt-go-chain-vm-v2/scenarios/fileresolver"
	mj "github.com/TerraDharitri/drt-go-chain-vm-v2/scenarios/json/model"
	"github.com/TerraDharitri/drt-go-chain-vm-v2/vmhost"
	"github.com/TerraDharitri/drt-go-chain-vm-v2/vmhost/hostCore"
)

var log = logger.GetOrCreate("vm/scenarios")

// TestVMType is the VM type argument we use in tests.
var TestVMType = []byte{0, 0}

// VMTestExecutor parses, interprets and executes both .test.json tests and .scen.json scenarios with VM.
type VMTestExecutor struct {
	World                 *worldhook.MockWorld
	vm                    vmi.VMExecutionHandler
	checkGas              bool
	scenGasScheduleLoaded bool
	fileResolver          fr.FileResolver
	exprReconstructor     er.ExprReconstructor
}

var _ mc.TestExecutor = (*VMTestExecutor)(nil)
var _ mc.ScenarioExecutor = (*VMTestExecutor)(nil)

// NewVMTestExecutor prepares a new VMTestExecutor instance.
func NewVMTestExecutor() (*VMTestExecutor, error) {
	world := worldhook.NewMockWorld()

	gasScheduleMap := config.MakeGasMapForTests()
	err := world.InitBuiltinFunctions(gasScheduleMap)
	if err != nil {
		return nil, err
	}

	blockGasLimit := uint64(10000000)
	vm, err := hostCore.NewVMHost(world, &vmhost.VMHostParameters{
		VMType:               TestVMType,
		BlockGasLimit:        blockGasLimit,
		GasSchedule:          gasScheduleMap,
		BuiltInFuncContainer: world.BuiltinFuncs.Container,
		ProtectedKeyPrefix:   []byte(core.ProtectedKeyPrefix),
		EnableEpochsHandler: &mock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == hostCore.SCDeployFlag || flag == hostCore.AheadOfTimeGasUsageFlag || flag == hostCore.RepairCallbackFlag || flag == hostCore.BuiltInFunctionsFlag
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &VMTestExecutor{
		World:                 world,
		vm:                    vm,
		checkGas:              true,
		scenGasScheduleLoaded: false,
		fileResolver:          nil,
		exprReconstructor:     er.ExprReconstructor{},
	}, nil
}

// GetVM yields a reference to the VMExecutionHandler used.
func (ae *VMTestExecutor) GetVM() vmi.VMExecutionHandler {
	return ae.vm
}

func (ae *VMTestExecutor) gasScheduleMapFromScenarios(scenGasSchedule mj.GasSchedule) (config.GasScheduleMap, error) {
	switch scenGasSchedule {
	case mj.GasScheduleDefault:
		return gasSchedules.LoadGasScheduleConfig(gasSchedules.GetV3())
	case mj.GasScheduleDummy:
		return config.MakeGasMapForTests(), nil
	case mj.GasScheduleV1:
		return gasSchedules.LoadGasScheduleConfig(gasSchedules.GetV1())
	case mj.GasScheduleV2:
		return gasSchedules.LoadGasScheduleConfig(gasSchedules.GetV2())
	case mj.GasScheduleV3:
		return gasSchedules.LoadGasScheduleConfig(gasSchedules.GetV3())
	default:
		return nil, fmt.Errorf("unknown scenario GasSchedule: %d", scenGasSchedule)
	}
}

// SetScenariosGasSchedule updates the gas costs based on the scenario config
// only changes the gas schedule once,
// this prevents subsequent gasSchedule declarations in externalSteps to overwrite
func (ae *VMTestExecutor) SetScenariosGasSchedule(newGasSchedule mj.GasSchedule) error {
	if ae.scenGasScheduleLoaded {
		return nil
	}
	gasSchedule, err := ae.gasScheduleMapFromScenarios(newGasSchedule)
	if err != nil {
		return err
	}
	ae.scenGasScheduleLoaded = true
	ae.vm.GasScheduleChange(gasSchedule)
	return nil
}
