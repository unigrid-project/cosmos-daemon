package types

const (
	// ModuleName defines the module name
	ModuleName = "pax"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_pax"
)

var (
	ParamsKey = []byte("p_pax_params")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
