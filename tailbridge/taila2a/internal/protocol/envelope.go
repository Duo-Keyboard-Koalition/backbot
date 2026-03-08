package protocol

import "encoding/json"

// Envelope wraps messages crossing the tailnet bridge/gateway boundary.
type Envelope struct {
	SourceNode string          `json:"source_node,omitempty"`
	DestNode   string          `json:"dest_node"`
	Payload    json.RawMessage `json:"payload"`
}
