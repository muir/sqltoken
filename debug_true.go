//go:build debugSQLToken

package sqltoken

const debug = true

type DebugExtra struct {
	Debug string
}

func (t *Token) SetDebug(s string) {
	t.DebugExtra.Debug = s
}

func (t *Token) Debug() string {
	return t.DebugExtra.Debug
}
