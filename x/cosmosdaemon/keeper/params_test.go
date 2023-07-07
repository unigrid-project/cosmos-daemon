package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "github.com/unigrid-project/cosmos-daemon/testutil/keeper"
	"github.com/unigrid-project/cosmos-daemon/x/cosmosdaemon/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.CosmosdaemonKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
