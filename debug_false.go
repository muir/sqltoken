//go:build !debugSQLToken

package sqltoken

type DebugExtra struct{}

const debug = false

func (t *Token) SetDebug(_ string) {}
func (t *Token) Debug() string     { return "" }
