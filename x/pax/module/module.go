package pax

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"

	"github.com/unigrid-project/cosmos-common/common/httpclient"
	// this line is used by starport scaffolding # 1

	modulev1 "pax/api/pax/pax/module"
	"pax/x/pax/keeper"
	"pax/x/pax/types"
)

var (
	_ module.AppModuleBasic      = (*AppModule)(nil)
	_ module.AppModuleSimulation = (*AppModule)(nil)
	_ module.HasGenesis          = (*AppModule)(nil)
	_ module.HasInvariants       = (*AppModule)(nil)
	_ module.HasConsensusVersion = (*AppModule)(nil)

	_ appmodule.AppModule       = (*AppModule)(nil)
	_ appmodule.HasBeginBlocker = (*AppModule)(nil)
	_ appmodule.HasEndBlocker   = (*AppModule)(nil)
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

type AppModuleBasic struct {
	cdc codec.BinaryCodec
}

func NewAppModuleBasic(cdc codec.BinaryCodec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

type AppModule struct {
	AppModuleBasic

	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
	}
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)

	InitGenesis(ctx, am.keeper, genState)
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genState)
}

func (AppModule) ConsensusVersion() uint64 { return 1 }

func (am AppModule) BeginBlock(_ context.Context) error {
	hedgehogUrl := viper.GetString("hedgehog.hedgehog_url") + "/minimum/version"
	fmt.Println("Constructed Hedgehog URL:", hedgehogUrl)

	hedgehogAvailable := isConnectedToHedgehog(hedgehogUrl)

	if !hedgehogAvailable {
		timer := time.NewTimer(2 * time.Hour)

		defer func() {
			if r := recover(); r != nil {
				select {
				case <-timer.C:
					panic(r)
				case available := <-checkHedgehogAvailability(hedgehogUrl, timer.C):
					if available {
						fmt.Println("Recovered from panic: Hedgehog server is now available.")
					} else {
						panic(r)
					}
				}
			}
		}()

		panic("Hedgehog is not available. Node is shutting down.")
	}

	fmt.Println("Hedgehog is available.")
	return nil
}

func (am AppModule) EndBlock(_ context.Context) error {
	return nil
}

func checkHedgehogAvailability(hedgehogUrl string, timerC <-chan time.Time) <-chan bool {
	availabilityChan := make(chan bool)
	go func() {
		defer close(availabilityChan)
		for {
			select {
			case <-timerC:
				fmt.Println("Timer expired in monitoring goroutine")
				availabilityChan <- false
				return
			default:
				if isConnectedToHedgehog(hedgehogUrl) {
					fmt.Println("Hedgehog server became available")
					availabilityChan <- true
					return
				}
				fmt.Println("Checking Hedgehog availability...")
				time.Sleep(5 * time.Second)
			}
		}
	}()
	return availabilityChan
}

func isConnectedToHedgehog(serverUrl string) bool {
	response, err := httpclient.Client.Get(serverUrl)

	if err != nil {
		fmt.Println("Error accessing hedgehog:", err.Error())
		return false
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err.Error())
		return false
	}

	if response.StatusCode != http.StatusOK {
		fmt.Printf("Received non-OK response from hedgehog server: %d\n", response.StatusCode)
		return false
	}

	var versionInfo struct {
		MinimumVersion struct {
			Version           string `json:"version"`
			HedgehogProtocol  string `json:"hedgehog_protocoll"`
			GridsporkProtocol string `json:"gridspork_protocoll"`
		} `json:"minimum_version"`
		CurrentVersion struct {
			Version           string `json:"version"`
			HedgehogProtocol  string `json:"hedgehog_protocoll"`
			GridsporkProtocol string `json:"gridspork_protocoll"`
		} `json:"current_version"`
	}

	if err := json.Unmarshal(body, &versionInfo); err != nil {
		fmt.Println("Error unmarshalling response body:", err.Error())
		return false
	}

	if versionInfo.MinimumVersion.Version == "" {
		fmt.Println("Minimum version not set, continuing...")
		return true
	}

	minVersion := strings.TrimSuffix(versionInfo.MinimumVersion.Version, "-SNAPSHOT")
	currVersion := strings.TrimSuffix(versionInfo.CurrentVersion.Version, "-SNAPSHOT")
	if strings.Contains(currVersion, "-BASTARD") {
		fmt.Println("Current hedgehog version contains an unauthorized build, rejecting version")
		return false
	}
	currVersion = strings.TrimSuffix(currVersion, "-BASTARD")

	if currVersion != minVersion {
		fmt.Printf("Current version %s does not match minimum version %s\n", currVersion, minVersion)
		return false
	}

	currHedgehogProtocol := strings.TrimPrefix(versionInfo.CurrentVersion.HedgehogProtocol, "hedgehog/")
	currGridsporkProtocol := strings.TrimPrefix(versionInfo.CurrentVersion.GridsporkProtocol, "gridspork/")

	if currHedgehogProtocol != versionInfo.MinimumVersion.HedgehogProtocol {
		fmt.Printf("Current hedgehog protocol %s does not match minimum hedgehog protocol %s\n", currHedgehogProtocol, versionInfo.MinimumVersion.HedgehogProtocol)
		return false
	}

	if currGridsporkProtocol != versionInfo.MinimumVersion.GridsporkProtocol {
		fmt.Printf("Current gridspork protocol %s does not match minimum gridspork protocol %s\n", currGridsporkProtocol, versionInfo.MinimumVersion.GridsporkProtocol)
		return false
	}

	fmt.Println("Hedgehog server is running the correct version.")
	return true
}

func (am AppModule) IsOnePerModuleType() {}

func (am AppModule) IsAppModule() {}

// ----------------------------------------------------------------------------
// App Wiring Setup
// ----------------------------------------------------------------------------

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	StoreService store.KVStoreService
	Cdc          codec.Codec
	Config       *modulev1.Module
	Logger       log.Logger

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	PaxKeeper keeper.Keeper
	Module    appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}
	k := keeper.NewKeeper(
		in.Cdc,
		in.StoreService,
		in.Logger,
		authority.String(),
	)
	m := NewAppModule(
		in.Cdc,
		k,
		in.AccountKeeper,
		in.BankKeeper,
	)

	return ModuleOutputs{PaxKeeper: k, Module: m}
}
