package keeper_test

import (
	"testing"

	testkeeper "github.com/unigrid-project/cosmos-daemon/testutil/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/unigrid-project/cosmos-daemon/x/cosmosdaemon/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.CosmosdaemonKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
