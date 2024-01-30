package app

import (
	"errors"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

type HandleOptions struct {
	ante.HandlerOptions

	WasmConfig *wasmTypes.WasmConfig
	WasmKepper *wasmkeeper.Keeper
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errors.New("account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return nil, errors.New("bank keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return nil, errors.New("sign mode handler is required for ante builder")
	}
	if options.WasmConfig == nil {
		return nil, errors.New("wasm config is required for ante builder")
	}
	if options.TXCounterStoreService == nil {
		return nil, errors.New("wasm store service is required for ante builder")
	}
	anteDecorators := []sdk.AnteDecorator{}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
