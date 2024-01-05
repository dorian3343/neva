package analyzer

import (
	"errors"
	"fmt"

	"github.com/nevalang/neva/internal/compiler"
	src "github.com/nevalang/neva/pkg/sourcecode"
	ts "github.com/nevalang/neva/pkg/typesystem"
)

var ErrStructFieldNotFound = errors.New("Struct field not found")

//nolint:funlen
func (a Analyzer) analyzeComponentNetwork(
	net []src.Connection,
	compInterface src.Interface,
	nodes map[string]src.Node,
	nodesIfaces map[string]src.Interface,
	scope src.Scope,
) ([]src.Connection, *compiler.Error) {
	nodesUsage := make(map[string]NodeNetUsage, len(nodes))

	for _, conn := range net {
		outportTypeExpr, err := a.getSenderType(conn.SenderSide, compInterface.IO.In, nodes, nodesIfaces, scope)
		if err != nil {
			return nil, compiler.Error{
				Location: &scope.Location,
				Meta:     &conn.SenderSide.Meta,
			}.Merge(err)
		}

		if conn.SenderSide.PortAddr != nil { // mark node's outport as used if sender isn't const ref
			senderNodeName := conn.SenderSide.PortAddr.Node
			outportName := conn.SenderSide.PortAddr.Port
			if _, ok := nodesUsage[senderNodeName]; !ok {
				nodesUsage[senderNodeName] = NodeNetUsage{
					In:  map[string]struct{}{},
					Out: map[string]struct{}{},
				}
			}
			nodesUsage[senderNodeName].Out[outportName] = struct{}{}
		}

		if len(conn.SenderSide.Selectors) > 0 {
			lastFieldType, err := ts.GetStructFieldTypeByPath(outportTypeExpr, conn.SenderSide.Selectors)
			if err != nil {
				return nil, &compiler.Error{
					Err:      err,
					Location: &scope.Location,
					Meta:     &conn.Meta,
				}
			}
			outportTypeExpr = lastFieldType
		}

		for _, receiver := range conn.ReceiverSides {
			inportTypeExpr, err := a.getReceiverType(receiver, compInterface.IO.Out, nodes, nodesIfaces, scope)
			if err != nil {
				return nil, compiler.Error{
					Err:      errors.New("Unable to get receiver type"),
					Location: &scope.Location,
					Meta:     &receiver.Meta,
				}.Merge(err)
			}

			if err := a.resolver.IsSubtypeOf(outportTypeExpr, inportTypeExpr, scope); err != nil {
				return nil, &compiler.Error{
					Err: fmt.Errorf(
						"Subtype checking failed: %v -> %v: %w",
						conn.SenderSide, receiver, err,
					),
					Location: &scope.Location,
					Meta:     &conn.Meta,
				}
			}

			// mark node's inport as used
			receiverNodeName := receiver.PortAddr.Node
			inportName := receiver.PortAddr.Port
			if _, ok := nodesUsage[receiverNodeName]; !ok {
				nodesUsage[receiverNodeName] = NodeNetUsage{
					In:  map[string]struct{}{}, // we don't use nodeIfaces for the same reason
					Out: map[string]struct{}{}, // as with outports
				}
			}
			nodesUsage[receiverNodeName].In[inportName] = struct{}{}
		}
	}

	if err := a.checkNodesUsage(nodesIfaces, scope, nodesUsage); err != nil {
		return nil, compiler.Error{Location: &scope.Location}.Merge(err)
	}

	return net, nil
}

// NodeNetUsage represents how network uses node's ports.
type NodeNetUsage struct {
	In, Out map[string]struct{}
}

// checkNodesUsage ensures that for every node we use all its inports
// and (if they are outports) at least one it's outport.
func (Analyzer) checkNodesUsage(
	nodesIfaces map[string]src.Interface,
	scope src.Scope,
	nodesUsage map[string]NodeNetUsage,
) *compiler.Error {
	for nodeName, nodeIface := range nodesIfaces {
		nodeUsage, ok := nodesUsage[nodeName]
		if !ok {
			return &compiler.Error{
				Err:      fmt.Errorf("%w: %v", ErrUnusedNode, nodeName),
				Location: &scope.Location,
			}
		}

		for inportName := range nodeIface.IO.In {
			if _, ok := nodeUsage.In[inportName]; !ok {
				meta := nodeIface.IO.In[inportName].Meta
				return &compiler.Error{
					Err:      fmt.Errorf("%w: node '%v', inport '%v'", ErrUnusedNodeInport, nodeName, inportName),
					Location: &scope.Location,
					Meta:     &meta,
				}
			}
		}

		if len(nodeIface.IO.Out) == 0 { // such components exist in stdlib, user cannot create them
			continue
		}

		atLeastOneOutportIsUsed := false
		for outportName := range nodeIface.IO.Out {
			if _, ok := nodeUsage.Out[outportName]; ok {
				atLeastOneOutportIsUsed = true
			}
		}
		if !atLeastOneOutportIsUsed {
			return &compiler.Error{
				Err:      fmt.Errorf("%w: %v", ErrUnusedNodeOutports, nodeName),
				Location: &scope.Location,
				Meta:     &nodeIface.Meta,
			}
		}
	}

	return nil
}

func (a Analyzer) getReceiverType(
	receiverSide src.ReceiverConnectionSide,
	outports map[string]src.Port,
	nodes map[string]src.Node,
	nodesIfaces map[string]src.Interface,
	scope src.Scope,
) (ts.Expr, *compiler.Error) {
	if receiverSide.PortAddr.Node == "in" {
		return ts.Expr{}, &compiler.Error{
			Err:      ErrWriteSelfIn,
			Location: &scope.Location,
			Meta:     &receiverSide.PortAddr.Meta,
		}
	}

	if receiverSide.PortAddr.Node == "out" {
		outport, ok := outports[receiverSide.PortAddr.Port]
		if !ok {
			return ts.Expr{}, &compiler.Error{
				Err:      ErrInportNotFound,
				Location: &scope.Location,
				Meta:     &receiverSide.PortAddr.Meta,
			}
		}
		return outport.TypeExpr, nil
	}

	nodeInportType, err := a.getNodeInportType(receiverSide.PortAddr, nodes, scope)
	if err != nil {
		return ts.Expr{}, compiler.Error{
			Location: &scope.Location,
			Meta:     &receiverSide.PortAddr.Meta,
		}.Merge(err)
	}

	return nodeInportType, nil
}

func (a Analyzer) getNodeInportType(
	portAddr src.PortAddr,
	nodes map[string]src.Node,
	scope src.Scope,
) (ts.Expr, *compiler.Error) {
	node, ok := nodes[portAddr.Node]
	if !ok {
		return ts.Expr{}, &compiler.Error{
			Err:      fmt.Errorf("%w '%v'", ErrNodeNotFound, portAddr.Node),
			Location: &scope.Location,
			Meta:     &portAddr.Meta,
		}
	}

	entity, location, err := scope.Entity(node.EntityRef)
	if err != nil {
		return ts.Expr{}, &compiler.Error{
			Err:      err,
			Location: &scope.Location,
			Meta:     &portAddr.Meta,
		}
	}

	var iface src.Interface
	if entity.Kind == src.ComponentEntity {
		iface = entity.Component.Interface
	} else { // we assume that nodes are already validated so if it's not component then it's interface
		iface = entity.Interface
	}

	typ, aerr := a.getResolvedPortType(
		iface.IO.In,
		iface.TypeParams.Params,
		portAddr,
		node,
		scope,
	)
	if aerr != nil {
		return ts.Expr{}, compiler.Error{
			Err:      fmt.Errorf("Unable to resolve '%v' port type", portAddr),
			Location: &location,
			Meta:     &portAddr.Meta,
		}.Merge(aerr)
	}

	return typ, nil
}

func (a Analyzer) getResolvedPortType(
	ports map[string]src.Port,
	params []ts.Param,
	portAddr src.PortAddr,
	node src.Node,
	scope src.Scope,
) (ts.Expr, *compiler.Error) {
	port, ok := ports[portAddr.Port]
	if !ok {
		return ts.Expr{}, &compiler.Error{
			Err:      fmt.Errorf("%w '%v'", ErrNodePortNotFound, portAddr),
			Location: &scope.Location,
			Meta:     &portAddr.Meta,
		}
	}

	_, frame, err := a.resolver.ResolveFrame(node.TypeArgs, params, scope)
	if err != nil {
		return ts.Expr{}, &compiler.Error{
			Err:      err,
			Location: &scope.Location,
			Meta:     &node.Meta,
		}
	}

	resolvedOutportType, err := a.resolver.ResolveExprWithFrame(port.TypeExpr, frame, scope)
	if err != nil {
		return ts.Expr{}, &compiler.Error{
			Err:      err,
			Location: &scope.Location,
			Meta:     &node.Meta,
		}
	}

	return resolvedOutportType, nil
}

func (a Analyzer) getSenderType(
	senderSide src.SenderConnectionSide,
	inports map[string]src.Port,
	nodes map[string]src.Node,
	nodesIfaces map[string]src.Interface,
	scope src.Scope,
) (ts.Expr, *compiler.Error) {
	if senderSide.PortAddr == nil && senderSide.ConstRef == nil {
		return ts.Expr{}, &compiler.Error{
			Err:      ErrSenderIsEmpty,
			Location: &scope.Location,
			Meta:     &senderSide.Meta,
		}
	}

	if senderSide.ConstRef != nil {
		expr, err := a.getResolvedConstType(*senderSide.ConstRef, scope)
		if err != nil {
			return ts.Expr{}, compiler.Error{
				Location: &scope.Location,
				Meta:     &senderSide.ConstRef.Meta,
			}.Merge(err)
		}
		return expr, nil
	}

	if senderSide.PortAddr.Node == "out" {
		return ts.Expr{}, &compiler.Error{
			Err:      ErrReadSelfOut,
			Location: &scope.Location,
			Meta:     &senderSide.PortAddr.Meta,
		}
	}

	if senderSide.PortAddr.Node == "in" {
		inport, ok := inports[senderSide.PortAddr.Port]
		if !ok {
			return ts.Expr{}, &compiler.Error{
				Err:      ErrInportNotFound,
				Location: &scope.Location,
				Meta:     &senderSide.PortAddr.Meta,
			}
		}
		return inport.TypeExpr, nil
	}

	nodeOutportType, err := a.getNodeOutportType(*senderSide.PortAddr, nodes, nodesIfaces, scope)
	if err != nil {
		return ts.Expr{}, compiler.Error{
			Err:      ErrInportNotFound,
			Location: &scope.Location,
			Meta:     &senderSide.PortAddr.Meta,
		}.Merge(err)
	}

	return nodeOutportType, nil
}

func (a Analyzer) getNodeOutportType(
	portAddr src.PortAddr,
	nodes map[string]src.Node,
	nodesIfaces map[string]src.Interface,
	scope src.Scope,
) (ts.Expr, *compiler.Error) {
	node, ok := nodes[portAddr.Node]
	if !ok {
		return ts.Expr{}, &compiler.Error{
			Err:      fmt.Errorf("%w: %v", ErrNodeNotFound, portAddr.Node),
			Location: &scope.Location,
			Meta:     &portAddr.Meta,
		}
	}

	nodeIface, ok := nodesIfaces[portAddr.Node]
	if !ok {
		return ts.Expr{}, &compiler.Error{
			Err:      fmt.Errorf("%w: %v", ErrNodeNotFound, portAddr.Node),
			Location: &scope.Location,
			Meta:     &portAddr.Meta,
		}
	}

	typ, err := a.getResolvedPortType(
		nodeIface.IO.Out,
		nodeIface.TypeParams.Params,
		portAddr,
		node,
		scope,
	)
	if err != nil {
		return ts.Expr{}, &compiler.Error{
			Err:      fmt.Errorf("get resolved outport type: %w", err),
			Location: &scope.Location,
			Meta:     &portAddr.Meta,
		}
	}

	return typ, err
}

func (a Analyzer) getResolvedConstType(ref src.EntityRef, scope src.Scope) (ts.Expr, *compiler.Error) {
	entity, location, err := scope.Entity(ref)
	if err != nil {
		return ts.Expr{}, &compiler.Error{
			Err:      err,
			Location: &scope.Location,
			Meta:     &ref.Meta,
		}
	}

	if entity.Kind != src.ConstEntity {
		return ts.Expr{}, &compiler.Error{
			Err:      fmt.Errorf("%w: %v", errors.New("Entity found but is not constant"), entity.Kind),
			Location: &location,
			Meta:     entity.Meta(),
		}
	}

	if entity.Const.Ref != nil {
		expr, err := a.getResolvedConstType(*entity.Const.Ref, scope)
		if err != nil {
			return ts.Expr{}, compiler.Error{
				Location: &location,
				Meta:     &entity.Const.Meta,
			}.Merge(err)
		}
		return expr, nil
	}

	resolvedExpr, err := a.resolver.ResolveExpr(entity.Const.Value.TypeExpr, scope)
	if err != nil {
		return ts.Expr{}, &compiler.Error{
			Err:      err,
			Location: &scope.Location,
			Meta:     &entity.Const.Value.Meta,
		}
	}

	return resolvedExpr, nil
}