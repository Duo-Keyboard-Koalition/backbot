package main

import "github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/protocol"

func normalizeOutboundEnvelope(env *protocol.Envelope, bridgeName string) error {
	if env.DestNode == "" || len(env.Payload) == 0 {
		return errOutboundRequired
	}
	if env.SourceNode == "" {
		env.SourceNode = bridgeName
	}
	return nil
}

func validateInboundEnvelope(env *protocol.Envelope) error {
	if len(env.Payload) == 0 {
		return errInboundRequired
	}
	return nil
}
