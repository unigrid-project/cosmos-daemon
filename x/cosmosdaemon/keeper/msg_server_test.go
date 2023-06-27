package keeper_test

import (
	"context"
	"testing"

	keepertest "cosmos-daemon/testutil/keeper"
	"cosmos-daemon/x/cosmosdaemon/keeper"
	"cosmos-daemon/x/cosmosdaemon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.CosmosdaemonKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}

func TestMsgServer(t *testing.T) {
	ms, ctx := setupMsgServer(t)
	require.NotNil(t, ms)
	require.NotNil(t, ctx)
}
