package builtin

import (
	"errors"
	"testing"

	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/assert"
)

func TestCons(t *testing.T) {
	t.Parallel()

	table := []struct {
		title   string
		first   core.Any
		rest    core.Seq
		items   []core.Any
		wantSz  int
		wantErr error
	}{
		{
			title:  "NilSeq",
			first:  Int64(100),
			rest:   nil,
			wantSz: 1,
		},
		{
			title:  "ZeroLenSeq",
			first:  Int64(100),
			rest:   NewList(),
			wantSz: 1,
		},
		{
			title:  "OneItemSeq",
			first:  Int64(100),
			rest:   NewList(1),
			wantSz: 2,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			seq, err := Cons(tt.first, tt.rest)
			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr),
					"wantErr=%#v\ngot=%#v", tt.wantErr, err)
				assert.Nil(t, seq,
					"failed call to Cons() returned non-nil value")
			} else {
				count, err := seq.Count()
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSz, count)
			}
		})
	}
}
