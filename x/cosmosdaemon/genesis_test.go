package cosmosdaemon_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "github.com/unigrid-project/cosmos-daemon/testutil/keeper"
	"github.com/unigrid-project/cosmos-daemon/testutil/nullify"
	"github.com/unigrid-project/cosmos-daemon/x/cosmosdaemon"
	"github.com/unigrid-project/cosmos-daemon/x/cosmosdaemon/types"
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
