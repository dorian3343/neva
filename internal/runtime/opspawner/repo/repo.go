package repo

import (
	"errors"
	"fmt"
	"plugin"

	"github.com/emil14/neva/internal/core"
	"github.com/emil14/neva/internal/runtime"
)

var (
	ErrUnknownPkg   = errors.New("operator refers to unknown package")
	ErrPluginOpen   = errors.New("plugin could not be loaded")
	ErrPluginLookup = errors.New("exported entity not found")
	ErrTypeMismatch = errors.New("exported entity doesn't match operator signature")
	ErrOpNotFound   = errors.New("package has not implemented the operator")
)

type Plugin struct {
	pkgs  map[string]PluginData
	cache map[runtime.OperatorRef]func(core.IO) error
}

func (r Plugin) Operator(ref runtime.OperatorRef) (func(core.IO) error, error) {
	return func(io core.IO) error { // FIXME
		kick, err := io.In.Port("kick")
		if err != nil {
			return err
		}

		str, err := io.In.Port("msg")
		if err != nil {
			return err
		}

		go func() {
			for {
				<-kick
				msg := <-str
				fmt.Println(msg.Str())
			}
		}()

		return nil
	}, nil

	if op, ok := r.cache[ref]; ok {
		return op, nil
	}

	pluginData, ok := r.pkgs[ref.Pkg]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownPkg, ref.Pkg)
	}

	plug, err := plugin.Open(pluginData.Filepath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPluginOpen, err)
	}

	for _, export := range pluginData.Exports {
		sym, err := plug.Lookup(export)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPluginLookup, err)
		}

		op, ok := sym.(func(core.IO) error)
		if !ok {
			return nil, fmt.Errorf("%w: %T", ErrTypeMismatch, op)
		}

		r.cache[ref] = op
	}

	op, ok := r.cache[ref]
	if !ok {
		return nil, fmt.Errorf("%w: %v", ErrOpNotFound, ref)
	}

	return op, nil
}

type PluginData struct {
	Filepath string
	Exports  []string
}

func NewPlugin(pkgs map[string]PluginData) Plugin {
	return Plugin{
		pkgs: pkgs,
		cache: make(
			map[runtime.OperatorRef]func(core.IO) error,
			len(pkgs),
		),
	}
}