package keeper

import (
	"cosmos-daemon/x/cosmosdaemon/types"
)

var _ types.QueryServer = Keeper{}
