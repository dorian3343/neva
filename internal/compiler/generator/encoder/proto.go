package encoder

import (
	"errors"
	"fmt"

	"github.com/emil14/neva/internal/pkg/initutils"
	"github.com/emil14/neva/internal/runtime/src"
	"github.com/emil14/neva/pkg/runtimesdk"
)

type (
	Marshaler interface {
		Marshal(*runtimesdk.Program) ([]byte, error)
	}
	Caster interface {
		Cast(src.Program) (runtimesdk.Program, error)
	}
)

var (
	ErrCast    = errors.New("caster")
	ErrMarshal = errors.New("marshaller")
)

type Proto struct {
	marshaler Marshaler
	caster    Caster
}

func (p Proto) Encode(prog src.Program) ([]byte, error) {
	sdkProg, err := p.caster.Cast(prog)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCast, err)
	}

	bb, err := p.marshaler.Marshal(&sdkProg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMarshal, err)
	}

	return bb, nil
}

func MustNew(marshaler Marshaler, caster Caster) Proto {
	initutils.NilPanic(marshaler)
	return Proto{marshaler, caster}
}
