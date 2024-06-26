package funcs

import (
	"context"

	"github.com/nevalang/neva/internal/runtime"
	"golang.org/x/exp/slices"
)

type listSortFloat struct{}

func (p listSortFloat) Create(io runtime.FuncIO, _ runtime.Msg) (func(ctx context.Context), error) {
	dataIn, err := io.In.Port("data")
	if err != nil {
		return nil, err
	}

	resOut, err := io.Out.Port("res")
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		var data runtime.Msg

		for {
			select {
			case <-ctx.Done():
				return
			case data = <-dataIn:
			}

			clone := slices.Clone(data.List())
			slices.SortFunc(clone, func(i, j runtime.Msg) bool {
				return i.Float() < j.Float()
			})

			select {
			case <-ctx.Done():
				return
			case resOut <- runtime.NewListMsg(clone...):
			}
		}
	}, nil
}
