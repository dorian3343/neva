package analyzer

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/exp/maps"

	src "github.com/nevalang/neva/internal/compiler/sourcecode"
	ts "github.com/nevalang/neva/pkg/typesystem"
)

var (
	ErrEmptyProgram      = errors.New("Program cannot be empty")
	ErrMainPkgNotFound   = errors.New("Main package is not found")
	ErrEmptyPkg          = errors.New("Package cannot be empty")
	ErrUnknownEntityKind = errors.New("Entity kind can only be either component, interface, type of constant")
)

type Analyzer struct {
	resolver ts.Resolver
}

func (a Analyzer) AnalyzeExecutable(mod src.Module, mainPkgName string) (src.Module, error) {
	rootLocation := &src.Location{
		ModuleName: "entry",
		PkgName:    mainPkgName,
	}

	if _, ok := mod.Packages[mainPkgName]; !ok {
		return src.Module{}, Error{
			Err:      fmt.Errorf("%w: main package name '%s'", ErrMainPkgNotFound, mainPkgName),
			Location: rootLocation,
		}
	}

	if err := a.mainSpecificPkgValidation(mainPkgName, mod); err.Err != nil {
		return src.Module{}, Error{Location: rootLocation}.Merge(err)
	}

	analyzedPkgs, err := a.Analyze(mod)
	if err != nil {
		aerr := err.(Error) //nolint:forcetypeassert
		return src.Module{}, Error{Location: rootLocation}.Merge(&aerr)
	}

	return src.Module{
		Manifest: mod.Manifest,
		Packages: analyzedPkgs,
	}, nil
}

func (a Analyzer) Analyze(mod src.Module) (map[string]src.Package, error) {
	if len(mod.Packages) == 0 { // Analyze can be called directly so we need to check emptiness here
		return nil, Error{Err: ErrEmptyProgram}
	}

	pkgsCopy := make(map[string]src.Package, len(mod.Packages))
	maps.Copy(pkgsCopy, mod.Packages)
	modCopy := src.Module{
		Manifest: mod.Manifest, // readonly
		Packages: pkgsCopy,
	}

	for pkgName := range pkgsCopy {
		resolvedPkg, err := a.analyzePkg(pkgName, modCopy)
		if err != nil {
			return nil, Error{
				Location: &src.Location{
					ModuleName: "entry",
					PkgName:    pkgName,
				},
			}.Merge(err)
		}
		pkgsCopy[pkgName] = resolvedPkg
	}

	return pkgsCopy, nil
}

func (a Analyzer) analyzePkg(pkgName string, mod src.Module) (src.Package, *Error) {
	if len(pkgName) == 0 {
		return nil, &Error{
			Err: ErrEmptyPkg,
			Location: &src.Location{
				ModuleName: "entry",
				PkgName:    pkgName,
			},
		}
	}

	resolvedPkg := make(map[string]src.File, len(pkgName))
	for fileName, file := range mod.Packages[pkgName] {
		resolvedPkg[fileName] = src.File{
			Imports:  file.Imports,
			Entities: make(map[string]src.Entity, len(file.Entities)),
		}
	}

	if err := mod.Packages[pkgName].Entities(func(entity src.Entity, entityName, fileName string) error {
		scope := src.Scope{
			Module: mod,
			Location: src.Location{
				PkgName:  pkgName,
				FileName: fileName,
			},
		}

		resolvedEntity, err := a.analyzeEntity(entity, scope)
		if err != nil {
			return &Error{
				Err: err,
				Location: &src.Location{
					ModuleName: "entry",
					PkgName:    pkgName,
					FileName:   fileName,
				},
			}
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
	isStd := strings.HasPrefix(scope.Location.PkgName, "std/")

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
