package decoder

import (
	"github.com/emil14/neva/internal/runtime/src"
	"github.com/emil14/neva/pkg/runtimesdk"
)

type caster struct{}

func (c caster) Cast(in *runtimesdk.Program) (src.Program, error) {
	ports := c.castPorts(in)
	connections := c.castConnections(in)
	funcs := c.castOperators(in)
	constants := c.castConstants(in)
	triggers := c.castTriggers(in)

	return src.Program{
		Ports: ports,
		Net:   connections,
		Effects: src.Effects{
			Operator: funcs,
			Giver:    constants,
			Triggers: triggers,
		},
		StartPortAddr: src.Ports{
			Path: in.StartPort.Path,
			Port: in.StartPort.Port,
			Idx:  uint8(in.StartPort.Idx),
		},
	}, nil
}

func (c caster) castConstants(in *runtimesdk.Program) map[src.Ports]src.Msg {
	constants := make(map[src.Ports]src.Msg, len(in.Effects.Constants))

	for _, constant := range in.Effects.Constants {
		addr := src.Ports{
			Path: constant.OutPortAddr.Path,
			Port: constant.OutPortAddr.Port,
			Idx:  uint8(constant.OutPortAddr.Idx),
		}
		constants[addr] = c.castMsg(constant.Msg)
	}

	return constants
}

func (c caster) castTriggers(in *runtimesdk.Program) []src.TriggerNode {
	triggers := make([]src.TriggerNode, 0, len(in.Effects.Triggers))

	for _, sdkTrigger := range in.Effects.Triggers {
		srcTrigger := src.TriggerNode{
			Msg: c.castMsg(sdkTrigger.Msg),
		}

		srcTrigger.In = src.Ports{
			Path: sdkTrigger.InPortAddr.Path,
			Port: sdkTrigger.InPortAddr.Port,
			Idx:  uint8(sdkTrigger.InPortAddr.Idx),
		}
		srcTrigger.Out = src.Ports{
			Path: sdkTrigger.OutPortAddr.Path,
			Port: sdkTrigger.OutPortAddr.Port,
			Idx:  uint8(sdkTrigger.OutPortAddr.Idx),
		}

		triggers = append(triggers, srcTrigger)
	}

	return triggers
}

func (caster) castOperators(in *runtimesdk.Program) []src.OpRef {
	operators := make([]src.OpRef, 0, len(in.Effects.Operators))
	for _, operator := range in.Effects.Operators {
		inAddrs := make([]src.Ports, 0, len(operator.InPortAddrs))
		for _, addr := range operator.InPortAddrs {
			inAddrs = append(inAddrs, src.Ports{
				Path: addr.Path,
				Port: addr.Port,
				Idx:  uint8(addr.Idx),
			})
		}

		outAddrs := make([]src.Ports, 0, len(operator.OutPortAddrs))
		for _, addr := range operator.OutPortAddrs {
			outAddrs = append(outAddrs, src.Ports{
				Path: addr.Path,
				Port: addr.Port,
				Idx:  uint8(addr.Idx),
			})
		}

		operators = append(operators, src.OpRef{
			FuncRef: src.FuncRef{
				Class: operator.Ref.Pkg,
				Name:  operator.Ref.Name,
			},
			PortAddrs: src.OpPortAddrs{
				In:  inAddrs,
				Out: outAddrs,
			},
		})
	}
	return operators
}

func (caster) castConnections(in *runtimesdk.Program) []src.Connection {
	connections := make([]src.Connection, 0, len(in.Connections))
	for _, connection := range in.Connections {
		receivers := make([]src.ConnectionSide, 0, len(connection.ReceiverConnectionPoints))
		for _, receiver := range connection.ReceiverConnectionPoints {
			receivers = append(receivers, src.ConnectionSide{
				PortAddr: src.Ports{
					Path: receiver.InPortAddr.Path,
					Port: receiver.InPortAddr.Port,
					Idx:  uint8(receiver.InPortAddr.Idx),
				},
				Action:                        src.ConnectorAction(receiver.Type),
				TraverseDictPathActionPayload: receiver.StructFieldPath,
			})
		}
		connections = append(connections, src.Connection{
			SenderSide: src.Ports{
				Path: connection.SenderOutPortAddr.Path,
				Port: connection.SenderOutPortAddr.Port,
				Idx:  uint8(connection.SenderOutPortAddr.Idx),
			},
			ReceiverSides: receivers,
		})
	}
	return connections
}

func (caster) castPorts(in *runtimesdk.Program) map[src.Ports]uint8 {
	ports := make(map[src.Ports]uint8, len(in.Ports))

	for _, p := range in.Ports {
		ports[src.Ports{
			Path: p.Addr.Path,
			Port: p.Addr.Port,
			Idx:  uint8(p.Addr.Idx),
		}] = uint8(p.BufSize)
	}

	return ports
}

func (c caster) castMsg(in *runtimesdk.Msg) src.Msg { // add err?
	msg := src.Msg{}

	switch in.Type {
	case runtimesdk.MsgType_VALUE_TYPE_BOOL: //nolint // nosnakecase
		msg = src.Msg{
			Type: src.BoolMsg,
			Bool: in.Bool,
		}
	case runtimesdk.MsgType_VALUE_TYPE_INT: //nolint // nosnakecase
		msg = src.Msg{
			Type: src.IntMsg,
			Int:  int(in.Int),
		}
	case runtimesdk.MsgType_VALUE_TYPE_STR: //nolint // nosnakecase
		msg = src.Msg{
			Type: src.StrMsg,
			Str:  in.Str,
		}
	case runtimesdk.MsgType_VALUE_TYPE_STRUCT: //nolint // nosnakecase
		structMsg := make(map[string]src.Msg, len(in.Struct))
		for k, v := range in.Struct {
			structMsg[k] = c.castMsg(v)
		}
		msg = src.Msg{
			Type: src.StrMsg,
			Dict: structMsg,
		}
	}

	return msg
}

func NewCaster() caster {
	return caster{}
}