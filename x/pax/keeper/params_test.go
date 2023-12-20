package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "github.com/unigrid-project/pax/testutil/keeper"
	"github.com/unigrid-project/pax/x/pax/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.PaxKeeper(t)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}
