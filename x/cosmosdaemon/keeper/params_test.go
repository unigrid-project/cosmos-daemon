package keeper_test

import (
	"testing"

	testkeeper "cosmos-daemon/testutil/keeper"
	"cosmos-daemon/x/cosmosdaemon/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.CosmosdaemonKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
