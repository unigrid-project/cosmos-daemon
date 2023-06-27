package cosmosdaemon_test

import (
	"testing"

	keepertest "cosmos-daemon/testutil/keeper"
	"cosmos-daemon/testutil/nullify"
	"cosmos-daemon/x/cosmosdaemon"
	"cosmos-daemon/x/cosmosdaemon/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.CosmosdaemonKeeper(t)
	cosmosdaemon.InitGenesis(ctx, *k, genesisState)
	got := cosmosdaemon.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
