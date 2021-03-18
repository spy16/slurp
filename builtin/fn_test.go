package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spy16/slurp/core"
)

func TestFn_Invoke(t *testing.T) {
	t.Parallel()

	specimen := Fn{
		Env:  NewEnv(nil),
		Name: "foo",
		Funcs: []Func{
			{
				Variadic: false,
				Params:   []string{"arg0"},
				Body:     &ResolveExpr{"arg0"},
			},
		},
		Macro: false,
	}

	table := []struct {
		title   string
		args    []core.Any
		want    core.Any
		wantErr bool
	}{
		{
			title:   "InvalidArity",
			args:    []core.Any{},
			want:    nil,
			wantErr: true,
		},
		{
			title:   "Arity1Call",
			args:    []core.Any{Int64(100)},
			want:    Int64(100),
			wantErr: false,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got, err := specimen.Invoke(tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
