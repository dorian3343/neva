package validator

import (
	"fmt"

	"github.com/emil14/neva/internal/compiler/program"
)

type validator struct{}

func (v validator) Validate(mod program.Module) error {
	if err := v.validatePorts(mod.Interface()); err != nil {
		return err
	}

	if err := v.validateDeps(mod.Deps); err != nil {
		return err
	}

	if err := v.validateWorkers(mod.Deps, mod.Workers); err != nil {
		return err
	}

	// TODO check arr points - should be no holes

	return nil
}

// validatePorts checks that ports are not empty and there is no unknown
func (mod validator) validatePorts(io program.IO) error {
	if len(io.In) == 0 || len(io.Out) == 0 {
		return fmt.Errorf("ports len 0")
	}

	for port, t := range io.In {
		if t.Type == program.UnknownType {
			return fmt.Errorf("unknown type " + port)
		}
	}

	for port, t := range io.Out {
		if t.Type == program.UnknownType {
			return fmt.Errorf("unknown type " + port)
		}
	}

	return nil
}

// validateWorkers checks that every worker points to existing dependency.
func (v validator) validateWorkers(deps program.ComponentsIO, workers map[string]string) error {
	for workerName, depName := range workers {
		if _, ok := deps[depName]; !ok {
			return fmt.Errorf("invalid workers: worker '%s' points to unknown dependency '%s'", workerName, depName)
		}
	}

	return nil
}

// validateDeps validates ports of every dependency.
func (v validator) validateDeps(deps program.ComponentsIO) error {
	for name, dep := range deps {
		if err := v.validatePorts(dep); err != nil {
			return fmt.Errorf("invalid dep '%s': %w", name, err)
		}
	}

	return nil
}

func New() validator {
	return validator{}
}

func MustNew() validator {
	return New()
}

// // validate[]RelationsDef checks that all port connections are type safe.
// // Then it checks that all connections are wired in the right way so the program will not block.
// // Ports, dependencies and workers should be validated before passing here.
// func (v validator) validateNet(in Inports, out Outports, deps Deps, workers Workers, net Net) error {
// 	senderReceivers := Graph{}
// 	receiverSenders := Graph{}

// 	for sender, conns := range net {
// 		if sender == "out" {
// 			return errors.New("'out' node could not be sender")
// 		}

// 		var senderOutports Outports
// 		if sender == "in" {
// 			senderOutports = Outports(in)
// 		} else {
// 			senderOutports = deps[workers[sender]].Out
// 		}

// 		for outport, conn := range conns {
// 			senderPoint := PortPoint{Node: sender, Port: outport}
// 			senderOutport := ByName(senderOutports[outport])
// 			receivers := map[PortPoint]struct{}{}

// 			for receiver, inports := range conn {
// 				if receiver == "in" {
// 					return errors.New("'in' node could not be receiver")
// 				}

// 				var receiverInports Inports
// 				if sender == "out" {
// 					receiverInports = Inports(out)
// 				} else {
// 					receiverInports = Inports(deps[workers[sender]].Out)
// 				}

// 				for _, inport := range inports {
// 					receiverInport := ByName(receiverInports[inport])
// 					if senderOutport != receiverInport {
// 						return fmt.Errorf("mismatched types")
// 					}

// 					receiverPoint := PortPoint{Node: receiver, Port: inport}
// 					receivers[receiverPoint] = struct{}{}
// 					if _, ok := receiverSenders[receiverPoint]; !ok {
// 						receiverSenders[receiverPoint] = map[PortPoint]struct{}{}
// 					}

// 					receiverSenders[receiverPoint][senderPoint] = struct{}{}
// 				}
// 			}

// 			senderReceivers[senderPoint] = receivers
// 		}
// 	}

// 	if err := validateOutflow("in", in, out, deps, workers, senderReceivers); err != nil {
// 		return err
// 	}

// 	return validateInflow("out", in, out, deps, workers, senderReceivers)
// }

// func validateInflow(receiver string, in Inports, out Outports, deps Deps, workers Workers, graph Graph) error {
// 	return nil // TODO
// }

// // validateOutflow finds node and checks that all its inports are connected to some other nodes outports.
// // Then it does so recursively for every sender-node.
// func validateOutflow(sender string, in Inports, out Outports, deps Deps, workers Workers, graph Graph) error {
// 	var ports Ports
// 	if sender == "out" {
// 		ports = Ports(out)
// 	} else {
// 		depName := workers[sender]
// 		ports = Ports(deps[depName].In)
// 	}

// 	for port := range ports {
// 		points, ok := graph[PortPoint{Node: sender, Port: port}]
// 		if !ok {
// 			return fmt.Errorf("'%s' outport of '%s' node is not wired", port, sender)
// 		}
// 		for p := range points {
// 			if err := validateOutflow(p.Node, in, out, deps, workers, graph); err != nil { // TODO: cache?
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// // Graph maps receiver port with the list of its sender ports.
// type Graph map[PortPoint]map[PortPoint]struct{}

// type PortPoint struct {
// 	Node, Port string
// }