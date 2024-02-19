package main

import (
	"github.com/nevalang/neva/internal/compiler"
	"github.com/nevalang/neva/internal/compiler/analyzer"
	"github.com/nevalang/neva/internal/compiler/desugarer"
	"github.com/nevalang/neva/internal/compiler/irgen"
	"github.com/nevalang/neva/internal/compiler/parser"
	"github.com/nevalang/neva/internal/interpreter"
	"github.com/nevalang/neva/internal/pkgmanager"
	"github.com/nevalang/neva/internal/runtime"
	"github.com/nevalang/neva/internal/runtime/funcs"
	"github.com/nevalang/neva/pkg"
	"github.com/nevalang/neva/pkg/typesystem"
)

func newInterpreter() interpreter.Interpreter {
	// runtime
	connector := runtime.NewDefaultConnector()
	funcRunner := runtime.MustNewFuncRunner(funcs.CreatorRegistry())
	r := runtime.New(connector, funcRunner)

	// type-system
	terminator := typesystem.Terminator{}
	checker := typesystem.MustNewSubtypeChecker(terminator)
	resolver := typesystem.MustNewResolver(typesystem.Validator{}, checker, terminator)

	// compiler
	desugarer := desugarer.Desugarer{}
	analyzer := analyzer.MustNew(pkg.Version, resolver)
	irgen := irgen.New()
	prsr := parser.New(false)
	comp := compiler.New(
		prsr,
		desugarer,
		analyzer,
		irgen,
		nil, // we don't need backend for interpretation
	)

	// interpreter
	return interpreter.New(
		comp,
		interpreter.NewAdapter(),
		r,
		pkgmanager.New(
			"/Users/emil/projects/neva/std",
			"/Users/emil/projects/neva/thirdparty",
			prsr,
		),
	)
}
