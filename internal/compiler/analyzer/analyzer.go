package analyzer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nevalang/neva/internal/compiler/src"
	ts "github.com/nevalang/neva/pkg/typesystem"
	"golang.org/x/exp/maps"
)

type Analyzer struct {
	resolver ts.Resolver
}

func (a Analyzer) AnalyzeExecutable(prog src.Program, mainPkgName string) (src.Program, error) {
	if _, ok := prog[mainPkgName]; !ok {
		return src.Program{}, ErrMainPkgNotFound
	}

	if err := a.mainSpecificPkgValidation(mainPkgName, prog); err != nil {
		return src.Program{}, fmt.Errorf("main specific pkg validation: %w", err)
	}

	return a.Analyze(prog)
}

// FIXME if there's more than 1 main pkg it will not work
func (a Analyzer) Analyze(prog src.Program) (src.Program, error) {
	if len(prog) == 0 { // Analyze can be called directly so we need to check emptiness here
		return src.Program{}, ErrEmptyProgram
	}

	progCopy := make(src.Program, len(prog))
	maps.Copy(progCopy, prog)

	for pkgName := range progCopy {
		resolvedPkg, err := a.analyzePkg(pkgName, progCopy)
		if err != nil {
			return src.Program{}, fmt.Errorf("analyze pkg: %v: %w", pkgName, err)
		}
		progCopy[pkgName] = resolvedPkg
	}

	return progCopy, nil
}

var (
	ErrEmptyProgram      = errors.New("empty program")
	ErrMainPkgNotFound   = errors.New("main package not found")
	ErrEmptyPkg          = errors.New("package must not be empty")
	ErrUnknownEntityKind = errors.New("unknown entity kind")
)

// TODO check that there's no 2 entities with the same name
// and that there's no unused entities.
func (a Analyzer) analyzePkg(pkgName string, prog src.Program) (src.Package, error) {
	if len(pkgName) == 0 {
		return nil, ErrEmptyPkg
	}

	resolvedPkg := make(map[string]src.File, len(pkgName))
	for fileName, file := range prog[pkgName] {
		resolvedPkg[fileName] = src.File{
			Imports:  file.Imports,
			Entities: make(map[string]src.Entity, len(file.Entities)),
		}
	}

	if err := prog[pkgName].Entities(func(entity src.Entity, entityName, fileName string) error {
		scope := src.Scope{
			Prog: prog,
			Loc: src.ScopeLocation{
				PkgName:  pkgName,
				FileName: fileName,
			},
		}
		resolvedEntity, err := a.analyzeEntity(entity, scope)
		if err != nil {
			return fmt.Errorf("analyze entity: %v: %v: %w", entityName, fileName, err)
		}
		resolvedPkg[fileName].Entities[entityName] = resolvedEntity
		return nil
	}); err != nil {
		return nil, fmt.Errorf("entities: %w", err)
	}

	return resolvedPkg, nil
}

func (a Analyzer) analyzeEntity(entity src.Entity, scope src.Scope) (src.Entity, error) {
	resolvedEntity := src.Entity{
		Exported: entity.Exported,
		Kind:     entity.Kind,
	}
	isStd := strings.HasPrefix(scope.Loc.PkgName, "std/")

	switch entity.Kind {
	case src.TypeEntity:
		resolvedTypeDef, err := a.analyzeTypeDef(entity.Type, scope, analyzeTypeDefParams{isStd})
		if err != nil {
			return src.Entity{}, fmt.Errorf("resolve type: %w", err)
		}
		resolvedEntity.Type = resolvedTypeDef
	case src.ConstEntity:
		resolvedConst, err := a.analyzeConst(entity.Const, scope)
		if err != nil {
			return src.Entity{}, fmt.Errorf("analyze const: %w", err)
		}
		resolvedEntity.Const = resolvedConst
	case src.InterfaceEntity:
		resolvedInterface, err := a.analyzeInterface(entity.Interface, scope, analyzeInterfaceParams{
			allowEmptyInports:  false,
			allowEmptyOutports: false,
		})
		if err != nil {
			return src.Entity{}, fmt.Errorf("analyze interface: %w", err)
		}
		resolvedEntity.Interface = resolvedInterface
	case src.ComponentEntity:
		resolvedComp, err := a.analyzeComponent(entity.Component, scope, analyzeComponentParams{
			iface: analyzeInterfaceParams{
				allowEmptyInports:  isStd, // e.g. `Const` component has no inports
				allowEmptyOutports: isStd, // e.g. `Void` component has no outports
			},
		})
		if err != nil {
			return src.Entity{}, fmt.Errorf("analyze component: %w", err)
		}
		resolvedEntity.Component = resolvedComp
	default:
		return src.Entity{}, fmt.Errorf("%w: %v", ErrUnknownEntityKind, entity.Kind)
	}

	return resolvedEntity, nil
}

func MustNew(resolver ts.Resolver) Analyzer {
	return Analyzer{
		resolver: resolver,
	}
}
