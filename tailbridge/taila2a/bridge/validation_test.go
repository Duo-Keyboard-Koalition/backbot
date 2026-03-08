package main

import (
	"testing"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/protocol"
)

func TestNormalizeOutboundEnvelope(t *testing.T) {
	tests := []struct {
		name       string
		env        protocol.Envelope
		bridgeName string
		wantErr    bool
		wantSource string
	}{
		{
			name:       "fills source when missing",
			env:        protocol.Envelope{DestNode: "bridge-beta", Payload: []byte(`{"msg":"hi"}`)},
			bridgeName: "bridge-alpha",
			wantErr:    false,
			wantSource: "bridge-alpha",
		},
		{
			name:       "keeps explicit source",
			env:        protocol.Envelope{SourceNode: "custom", DestNode: "bridge-beta", Payload: []byte(`{"msg":"hi"}`)},
			bridgeName: "bridge-alpha",
			wantErr:    false,
			wantSource: "custom",
		},
		{
			name:       "errors when destination missing",
			env:        protocol.Envelope{Payload: []byte(`{"msg":"hi"}`)},
			bridgeName: "bridge-alpha",
			wantErr:    true,
		},
		{
			name:       "errors when payload missing",
			env:        protocol.Envelope{DestNode: "bridge-beta"},
			bridgeName: "bridge-alpha",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := tt.env
			err := normalizeOutboundEnvelope(&env, tt.bridgeName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeOutboundEnvelope() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && env.SourceNode != tt.wantSource {
				t.Fatalf("normalizeOutboundEnvelope() source = %q, want %q", env.SourceNode, tt.wantSource)
			}
		})
	}
}

func TestValidateInboundEnvelope(t *testing.T) {
	if err := validateInboundEnvelope(&protocol.Envelope{Payload: []byte(`{"ok":true}`)}); err != nil {
		t.Fatalf("validateInboundEnvelope() unexpected error: %v", err)
	}

	if err := validateInboundEnvelope(&protocol.Envelope{}); err == nil {
		t.Fatalf("validateInboundEnvelope() expected error for empty payload")
	}
}
