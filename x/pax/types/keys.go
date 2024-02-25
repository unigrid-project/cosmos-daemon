package types

const (
	// ModuleName defines the module name
	ModuleName = "pax"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_pax"
)

var (
	ParamsKey = []byte("p_pax")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
