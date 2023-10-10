package funcs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/nevalang/neva/internal/runtime"
)

func Read(ctx context.Context, io runtime.FuncIO) (func(), error) {
	sig, err := io.In.Port("sig")
	if err != nil {
		return nil, err
	}
	vout, err := io.Out.Port("v")
	if err != nil {
		return nil, err
	}
	return func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sig:
				select {
				case <-ctx.Done():
					return
				default:
					text, err := reader.ReadString('\n')
					if err != nil {
						panic(err) // TODO handle
					}
					select {
					case <-ctx.Done():
						return
					case vout <- runtime.NewStrMsg(text):
					}
				}
			}
		}
	}, nil
}

func Print(ctx context.Context, io runtime.FuncIO) (func(), error) {
	vin, err := io.In.Port("v")
	if err != nil {
		return nil, err
	}
	vout, err := io.Out.Port("v")
	if err != nil {
		return nil, err
	}
	return func() {
		for {
			select {
			case <-ctx.Done():
				return
			case v := <-vin:
				select {
				case <-ctx.Done():
					return
				default:
					fmt.Print(v.String())
					select {
					case <-ctx.Done():
						return
					case vout <- v:
					}
				}
			}
		}
	}, nil
}

func Lock(ctx context.Context, io runtime.FuncIO) (func(), error) {
	vin, err := io.In.Port("v")
	if err != nil {
		return nil, err
	}
	sig, err := io.In.Port("sig")
	if err != nil {
		return nil, err
	}
	vout, err := io.Out.Port("v")
	if err != nil {
		return nil, err
	}
	return func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-sig:
				select {
				case <-ctx.Done():
					return
				case v := <-vin:
					select {
					case <-ctx.Done():
						return
					case vout <- v:
					}
				}
			}
		}
	}, nil
}

func Const(ctx context.Context, io runtime.FuncIO) (func(), error) {
	msg := ctx.Value("msg")
	if msg == nil {
		return nil, errors.New("ctx msg not found")
	}

	v, ok := msg.(runtime.Msg)
	if !ok {
		return nil, errors.New("ctx value is not runtime message")
	}

	vout, err := io.Out.Port("v")
	if err != nil {
		return nil, err
	}

	return func() {
		for {
			select {
			case <-ctx.Done():
				return
			case vout <- v:
			}
		}
	}, nil
}

func Add(ctx context.Context, io runtime.FuncIO) (func(), error) {
	msg := ctx.Value("msg")
	if msg == nil {
		return nil, errors.New("ctx msg not found")
	}

	typ, ok := msg.(runtime.Msg)
	if !ok {
		return nil, errors.New("ctx value is not runtime message")
	}

	var handler func(a, b runtime.Msg) runtime.Msg
	switch typ.Int() {
	case 1: // int
		handler = func(a, b runtime.Msg) runtime.Msg {
			return runtime.NewIntMsg(a.Int() + b.Int())
		}
	case 2: // float
		handler = func(a, b runtime.Msg) runtime.Msg {
			return runtime.NewFloatMsg(a.Float() + b.Float())
		}
	case 3: // string
		handler = func(a, b runtime.Msg) runtime.Msg {
			return runtime.NewStrMsg(a.Str() + b.Str())
		}
	default:
		return nil, errors.New("unknown msg type")
	}

	a, err := io.In.Port("a")
	if err != nil {
		return nil, err
	}
	b, err := io.In.Port("b")
	if err != nil {
		return nil, err
	}
	vout, err := io.Out.Port("v")
	if err != nil {
		return nil, err
	}

	return func() {
		for {
			select {
			case <-ctx.Done():
				return
			case v1 := <-a:
				select {
				case <-ctx.Done():
					return
				case v2 := <-b:
					select {
					case <-ctx.Done():
						return
					case vout <- handler(v1, v2):
					}
				}
			}
		}
	}, nil
}

func ParseNum(ctx context.Context, io runtime.FuncIO) (func(), error) { //nolint:funlen
	msg := ctx.Value("msg")
	if msg == nil {
		return nil, errors.New("ctx msg not found")
	}

	typ, ok := msg.(runtime.Msg)
	if !ok {
		return nil, errors.New("ctx value is not runtime message")
	}

	vin, err := io.In.Port("v")
	if err != nil {
		return nil, err
	}

	vout, err := io.Out.Port("v")
	if err != nil {
		return nil, err
	}

	errout, err := io.Out.Port("err")
	if err != nil {
		return nil, err
	}

	var handler func(string) (runtime.Msg, error)
	if typ.Int() == 1 { // int
		handler = func(str string) (runtime.Msg, error) {
			v, err := strconv.Atoi(str)
			if err != nil {
				return nil, err
			}
			return runtime.NewIntMsg(int64(v)), nil
		}
	} else { // float
		handler = func(str string) (runtime.Msg, error) {
			v, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return nil, err
			}
			return runtime.NewFloatMsg(v), nil
		}
	}

	return func() {
		for {
			select {
			case <-ctx.Done():
				return
			case str := <-vin:
				v, err := handler(str.Str())
				if err != nil {
					select {
					case <-ctx.Done():
						return
					case errout <- runtime.NewStrMsg(err.Error()):
					}
					continue
				}
				select {
				case <-ctx.Done():
					return
				case vout <- v:
				}
			}
		}
	}, nil
}

func Repo() map[string]runtime.Func {
	return map[string]runtime.Func{
		"Read":     Read,
		"Print":    Print,
		"Lock":     Lock,
		"Const":    Const,
		"Add":      Add,
		"ParseNum": ParseNum,
	}
}