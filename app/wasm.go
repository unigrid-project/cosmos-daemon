package app

import (
	"fmt"
	"path/filepath"
	"strings"

	storetypes "cosmossdk.io/store/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibcfee "github.com/cosmos/ibc-go/v8/modules/apps/29-fee"
	ibcporttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	paxwasm "github.com/unigrid-project/cosmos-daemon/app/wasm"
)

func (app *App) registerWasmModules(appOpts servertypes.AppOptions) {
	if err := app.RegisterStores(); err != nil {
		panic(err)
	}
	keys := storetypes.NewKVStoreKeys(
		wasmtypes.StoreKey,
	)

	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	var wasmOpts []wasmkeeper.Option
	if cast.ToBool(appOpts.Get("telemetry.enabled")) {
		wasmOpts = append(wasmOpts, wasmkeeper.WithVMCacheMetrics(prometheus.DefaultRegisterer))
	}
	//add if we use custom commands in wasm
	//wasmOpts = append(wasmOpts, wasmkeeper.WithQueryPlugins(okp4wasm.CustomQueryPlugins(&app.LogicKeeper)))

	availableCapabilities := strings.Join(paxwasm.AllCapabilities(), ",")
	app.WasmKeeper = wasmkeeper.NewKeeper(
		app.AppCodec(),
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCFeeKeeper, // ISC4 Wrapper: fee IBC middleware
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		scopedWasmKeeper,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		availableCapabilities,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	var wasmStack ibcporttypes.IBCModule
	wasmStack = wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCFeeKeeper)
	wasmStack = ibcfee.NewIBCMiddleware(wasmStack, app.IBCFeeKeeper)

	ibcRouter := ibcporttypes.NewRouter()
	ibcRouter.AddRoute(wasmtypes.ModuleName, wasmStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	if err := app.RegisterModules(
		wasm.NewAppModule(app.appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper,
			app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
	); err != nil {
		panic(err)
	}
}

func AddWASMModuleManager(moduleManager module.BasicManager) {
	moduleManager[wasmtypes.ModuleName] = wasm.AppModule{}
	moduleManager[capabilitytypes.ModuleName] = capability.AppModule{}
}
