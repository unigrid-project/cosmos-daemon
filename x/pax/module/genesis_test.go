package pax_test

import (
	"testing"

	keepertest "pax/testutil/keeper"
	"pax/testutil/nullify"
	pax "pax/x/pax/module"
	"pax/x/pax/types"

	"github.com/stretchr/testify/require"
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
