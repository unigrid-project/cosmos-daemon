package pax

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	// this line is used by starport scaffolding # 1

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/unigrid-project/cosmos-sdk-common/common/httpclient"
	"github.com/unigrid-project/pax/x/pax/client/cli"
	"github.com/unigrid-project/pax/x/pax/keeper"
	"github.com/unigrid-project/pax/x/pax/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface that defines the independent methods a Cosmos SDK module needs to implement.
type AppModuleBasic struct {
	cdc codec.BinaryCodec
}

func NewAppModuleBasic(cdc codec.BinaryCodec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// Name returns the name of the module as a string
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the amino codec for the module, which is used to marshal and unmarshal structs to/from []byte in order to persist them in the module's KVStore
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterCodec(cdc)
}

// RegisterInterfaces registers a module's interface types and their concrete implementations as proto.Message
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns a default GenesisState for the module, marshalled to json.RawMessage. The default GenesisState need to be defined by the module developer and is primarily used for testing
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis used to validate the GenesisState, given in its json.RawMessage form
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// GetTxCmd returns the root Tx command for the module. The subcommands of this root command are used by end-users to generate new transactions containing messages defined in the module
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the root query command for the module. The subcommands of this root command are used by end-users to generate new queries to the subset of the state defined by the module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(types.StoreKey)
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface that defines the inter-dependent methods that modules need to implement
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

// RegisterServices registers a gRPC query service to respond to the module-specific gRPC queries
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// RegisterInvariants registers the invariants of the module. If an invariant deviates from its predicted value, the InvariantRegistry triggers appropriate logic (most often the chain will be halted)
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs the module's genesis initialization. It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	InitGenesis(ctx, am.keeper, genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion is a sequence number for state-breaking change of the module. It should be incremented on each consensus-breaking change introduced by the module. To avoid wrong/empty versions, the initial version should be set to 1
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock contains the logic that is automatically triggered at the beginning of each block
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	hedgehogUrl := viper.GetString("hedgehog.hedgehog_url") + "/gridspork"
	hedgehogAvailable := isConnectedToHedgehog(hedgehogUrl)

	if !hedgehogAvailable {
		// Start a timer
		timer := time.NewTimer(2 * time.Hour) // 2 hours until we panic and shutdown the node

		// Set up deferred panic handling
		defer func() {
			if r := recover(); r != nil {
				// Wait for either the timer to expire or the server to become available
				select {
				case <-timer.C:
					// Timer expired, re-panic
					panic(r)
				case available := <-checkHedgehogAvailability(hedgehogUrl, timer.C):
					if available {
						fmt.Println("Recovered from panic: Hedgehog server is now available.")
					} else {
						// Server did not become available in time, re-panic
						panic(r)
					}
				}
			}
		}()

		// Trigger panic
		panic("Hedgehog is not available. Node is shutting down.")
	}

	fmt.Println("Hedgehog is available.")
}

// checkHedgehogAvailability continuously checks if the Hedgehog server becomes available
// and returns a boolean value through a channel when the server is available or the timer expires.
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
				time.Sleep(5 * time.Second) // check interval, adjust as needed
			}
		}
	}()
	return availabilityChan
}

// EndBlock contains the logic that is automatically triggered at the end of each block
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// isConnectedToHedgehog performs an HTTP GET request to check the connectivity with the Hedgehog server.
func isConnectedToHedgehog(serverUrl string) bool {

	response, err := httpclient.Client.Get(serverUrl)

	if err != nil {
		fmt.Println("Error accessing hedgehog:", err.Error())
		return false
	}
	defer response.Body.Close()

	// Read and discard the response body
	_, err = io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err.Error())
		return false
	}

	// Check if the HTTP status is 200 OK
	if response.StatusCode == http.StatusOK {
		fmt.Printf("Received OK response from hedgehog server: %d\n", response.StatusCode)
		return true
	}

	fmt.Printf("Received non-OK response from hedgehog server: %d\n", response.StatusCode)

	return false
}
