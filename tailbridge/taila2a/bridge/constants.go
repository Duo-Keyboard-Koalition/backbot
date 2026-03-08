package main

import "errors"

const (
	defaultBridgeName  = "bridge-alpha"
	defaultLocalAgent  = "http://127.0.0.1:9090/api"
	defaultPeerInPort  = 8001
	defaultInboundPort = 8001
	defaultLocalListen = "127.0.0.1:8080"
	maxBodyBytes       = 1 << 20 // 1 MiB
)

var (
	errOutboundRequired = errors.New("dest_node and payload are required")
	errInboundRequired  = errors.New("payload is required")
)
