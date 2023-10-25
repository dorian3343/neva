package main

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (s Server) Initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	return protocol.InitializeResult{
		Capabilities: s.handler.CreateServerCapabilities(),
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    s.name,
			Version: &s.version,
		},
	}, nil
}

func (srv Server) Initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func (srv Server) Shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func (srv Server) SetTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}