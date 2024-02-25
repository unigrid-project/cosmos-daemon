package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "pax/testutil/keeper"
	"pax/x/pax/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := keepertest.PaxKeeper(t)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}
