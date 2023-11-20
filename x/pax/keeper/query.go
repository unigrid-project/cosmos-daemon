package keeper

import (
	"github.com/unigrid-project/pax/x/pax/types"
)

var _ types.QueryServer = Keeper{}
