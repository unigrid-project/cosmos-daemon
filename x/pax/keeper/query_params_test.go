package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "pax/testutil/keeper"
	"pax/x/pax/keeper"
	"pax/x/pax/types"
)

func TestParamsQuery(t *testing.T) {
	k, ctx := keepertest.PaxKeeper(t)
	qs := keeper.NewQueryServerImpl(k)
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params))

	response, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
