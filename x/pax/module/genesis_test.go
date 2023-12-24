package pax_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "github.com/unigrid-project/pax/testutil/keeper"
	"github.com/unigrid-project/pax/testutil/nullify"
	pax "github.com/unigrid-project/pax/x/pax/module"
	"github.com/unigrid-project/pax/x/pax/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.PaxKeeper(t)
	pax.InitGenesis(ctx, k, genesisState)
	got := pax.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
