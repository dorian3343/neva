package analyze

import (
	"errors"
	"fmt"

	"github.com/emil14/neva/internal/compiler/src"
)

var (
	ErrEntity              = errors.New("analyze entity")
	ErrScopeGetLocalEntity = errors.New("scope get local entity")
	ErrType                = errors.New("analyze type")
	ErrUnknownEntityKind   = errors.New("unknown entity kind")
	ErrMsg                 = errors.New("analyze msg")
	ErrComponent           = errors.New("analyze component")
	ErrInterface           = errors.New("analyze interface")
)

func (a Analyzer) analyzeEntities(pkg src.Pkg, scope Scope) (map[string]src.Entity, map[src.EntityRef]struct{}, error) {
	resolvedPkgEntities := make(map[string]src.Entity, len(pkg.Entities))
	allUsedEntities := map[src.EntityRef]struct{}{} // both local and imported

	for entityName, entity := range pkg.Entities {
		if entity.Exported || entityName == pkg.RootComponent {
			allUsedEntities[src.EntityRef{Name: entityName}] = struct{}{} // normalize?
		}

		resolvedEntity, entitiesUsedByEntity, err := a.analyzeEntity(entityName, scope)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %v", errors.Join(ErrEntity, err), entityName)
		}

		for entityRef := range entitiesUsedByEntity {
			allUsedEntities[entityRef] = struct{}{}
		}

		resolvedPkgEntities[entityName] = resolvedEntity
	}

	return resolvedPkgEntities, allUsedEntities, nil
}

func (a Analyzer) analyzeEntity(name string, scope Scope) (src.Entity, map[src.EntityRef]struct{}, error) {
	entity, err := scope.getLocalEntity(name)
	if err != nil {
		return src.Entity{}, nil, errors.Join(ErrScopeGetLocalEntity, err)
	}

	switch entity.Kind { // https://github.com/emil14/neva/issues/186
	case src.TypeEntity:
		resolvedDef, usedTypeEntities, err := a.analyzeType(name, scope)
		if err != nil {
			return src.Entity{}, nil, errors.Join(ErrType, err)
		}
		return src.Entity{
			Type:     resolvedDef,
			Kind:     src.TypeEntity,
			Exported: entity.Exported,
		}, usedTypeEntities, nil
	case src.MsgEntity:
		resolvedMsg, usedEntities, err := a.analyzeMsg(entity.Msg, scope, nil)
		if err != nil {
			return src.Entity{}, nil, errors.Join(ErrMsg, err)
		}
		return src.Entity{
			Msg:      resolvedMsg,
			Kind:     src.MsgEntity,
			Exported: entity.Exported,
		}, usedEntities, nil
	case src.InterfaceEntity:
		resolvedInterface, used, err := a.analyzeInterface(entity.Interface, scope)
		if err != nil {
			return src.Entity{}, nil, errors.Join(ErrInterface, err)
		}
		return src.Entity{
			Exported:  entity.Exported,
			Kind:      src.InterfaceEntity,
			Interface: resolvedInterface,
		}, used, nil
	case src.ComponentEntity:
		resolvedComponent, used, err := a.analyzeComponent(entity.Component, scope)
		if err != nil {
			return src.Entity{}, nil, errors.Join(ErrComponent, err)
		}
		return src.Entity{
			Exported:  entity.Exported,
			Kind:      src.ComponentEntity,
			Component: resolvedComponent,
		}, used, nil
	default:
		return src.Entity{}, nil, ErrUnknownEntityKind
	}

	return src.Entity{}, map[src.EntityRef]struct{}{}, nil
}