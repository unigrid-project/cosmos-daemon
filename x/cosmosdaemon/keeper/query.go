package keeper

import (
	"github.com/unigrid-project/cosmos-daemon/x/cosmosdaemon/types"
)

var _ types.QueryServer = Keeper{}
