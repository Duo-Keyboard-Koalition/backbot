package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"tailscale.com/tsnet"
)

func tsHTTPClient(srv *tsnet.Server, timeout time.Duration) *http.Client {
	tr := &http.Transport{}
	tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return srv.Dial(ctx, network, addr)
	}
	return &http.Client{Transport: tr, Timeout: timeout}
}

func logSelfTailscaleIPs(srv *tsnet.Server) {
	lc, err := srv.LocalClient()
	if err != nil {
		log.Printf("bridge tailscale local client unavailable: %v", err)
		return
	}

	for i := 0; i < 30; i++ {
		st, err := lc.Status(context.Background())
		if err != nil {
			log.Printf("bridge tailscale status unavailable: %v", err)
			return
		}

		if st.Self != nil && len(st.Self.TailscaleIPs) > 0 {
			ips := make([]string, 0, len(st.Self.TailscaleIPs))
			for _, ip := range st.Self.TailscaleIPs {
				ips = append(ips, ip.String())
			}
			log.Printf("bridge tailnet IP(s): %s", strings.Join(ips, ", "))
			return
		}

		if i == 0 {
			log.Printf("bridge tailnet IP pending: node not yet fully connected")
		}
		time.Sleep(2 * time.Second)
	}

	log.Printf("bridge tailnet IP still pending after retry window")
}
