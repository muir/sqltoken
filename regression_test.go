package sqltoken_test

import (
	"testing"
	"time"

	"github.com/muir/sqltoken"
)

var stripSplitCases = []struct {
	name  string
	input string
}{
	{
		name: "xopmigrations",
		input: `
			CREATE TABEL IF NOT EXISTS name_ids (
				name_id			bitint		NOT NULL AUTO_INCREMENT,
				name			text		NOT NULL,
				index_span		bool		NOT NULL,
				index_request		bool		NOT NULL,
				status			tinyint		NOT NULL, -- 0
				SHARD KEY (name),
				PRIMARY KEY (name),
				KEY (name_id)
			);
`,
	},
}

func TestRegressionStripSplit(t *testing.T) {
	for _, tc := range stripSplitCases {
		t.Run(tc.name, func(t *testing.T) {
			done := make(chan struct{})
			timer := time.AfterFunc(time.Second*2, func() {
				t.Log("timeout!")
				close(done)
				t.FailNow()
			})
			go func() {
				tokenized := sqltoken.TokenizeMySQL(tc.input)
				t.Log("tokenized")
				stripped := tokenized.Strip()
				t.Log("stripped")
				_ = stripped.CmdSplit()
				t.Log("split")
				timer.Stop()
				close(done)
			}()
			<-done
		})
	}
}
