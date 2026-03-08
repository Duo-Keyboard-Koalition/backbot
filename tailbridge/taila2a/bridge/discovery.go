package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

// AgentInfo represents a discovered agent on the tailnet
type AgentInfo struct {
	Name      string    `json:"name"`
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	Online    bool      `json:"online"`
	LastSeen  time.Time `json:"last_seen"`
	Gateways  []GatewayInfo `json:"gateways"`
}

// GatewayInfo represents an open gateway port on an agent
type GatewayInfo struct {
	Port      int    `json:"port"`
	Protocol  string `json:"protocol"`
	Service   string `json:"service"`
}

// DiscoveryService manages agent discovery on the tailnet
type DiscoveryService struct {
	srv           *tsnet.Server
	localClient   *tailscale.LocalClient
	agents        map[string]*AgentInfo
	mu            sync.RWMutex
	discoverTimer *time.Ticker
	stopChan      chan struct{}
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(srv *tsnet.Server) (*DiscoveryService, error) {
	lc, err := srv.LocalClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get local client: %w", err)
	}

	return &DiscoveryService{
		srv:     srv,
		localClient: lc,
		agents:  make(map[string]*AgentInfo),
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins periodic agent discovery (LEGACY - with port scanning)
// Deprecated: Use StartPassive for AIP-compatible mode
func (d *DiscoveryService) Start(interval time.Duration) {
	log.Printf("[discovery] starting agent discovery (interval: %v) - LEGACY MODE", interval)

	d.discoverTimer = time.NewTicker(interval)
	go d.runDiscovery()

	// Initial discovery
	go d.discoverOnce()
}

// StartPassive begins passive agent discovery (no port scanning)
// Only tracks peers from Tailscale status, doesn't scan ports
func (d *DiscoveryService) StartPassive(interval time.Duration) {
	log.Printf("[discovery] starting passive discovery (interval: %v) - AIP MODE", interval)

	d.discoverTimer = time.NewTicker(interval)
	go d.runDiscoveryPassive()

	// Initial discovery
	go d.discoverOncePassive()
}

// Stop halts the discovery service
func (d *DiscoveryService) Stop() {
	if d.discoverTimer != nil {
		d.discoverTimer.Stop()
	}
	close(d.stopChan)
}

func (d *DiscoveryService) runDiscovery() {
	for {
		select {
		case <-d.discoverTimer.C:
			d.discoverOnce()
		case <-d.stopChan:
			return
		}
	}
}

func (d *DiscoveryService) discoverOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	status, err := d.localClient.Status(ctx)
	if err != nil {
		log.Printf("[discovery] failed to get status: %v", err)
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Mark all agents as offline first
	for _, agent := range d.agents {
		agent.Online = false
	}

	// Discover active agents
	for _, peer := range status.Peer {
		if peer.DNSName == "" {
			continue
		}

		hostname := peer.DNSName
		if len(hostname) > 0 && hostname[len(hostname)-1] == '.' {
			hostname = hostname[:len(hostname)-1]
		}

		name := hostname
		if idx := indexOf(hostname, '.'); idx > 0 {
			name = hostname[:idx]
		}

		agent := d.agents[name]
		if agent == nil {
			agent = &AgentInfo{
				Name:     name,
				Hostname: hostname,
				Gateways: make([]GatewayInfo, 0),
			}
			d.agents[name] = agent
		}

		agent.IP = peer.TailscaleIPs[0].String()
		agent.Online = true
		agent.LastSeen = time.Now()

		// Scan for open gateway ports
		agent.Gateways = d.scanGateways(ctx, agent.IP)

		log.Printf("[discovery] found agent: %s (%s) - %d gateways", 
			name, agent.IP, len(agent.Gateways))
	}

	// Remove stale agents (offline for > 5 minutes)
	for name, agent := range d.agents {
		if !agent.Online && time.Since(agent.LastSeen) > 5*time.Minute {
			delete(d.agents, name)
			log.Printf("[discovery] removed stale agent: %s", name)
		}
	}
}

func (d *DiscoveryService) scanGateways(ctx context.Context, ip string) []GatewayInfo {
	gateways := make([]GatewayInfo, 0)
	
	// Common gateway ports to check
	ports := []int{8001, 8080, 9090, 5000, 3000, 80, 443, 8443}
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			
			connCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			
			conn, err := (&net.Dialer{}).DialContext(connCtx, "tcp", fmt.Sprintf("%s:%d", ip, p))
			if err == nil {
				conn.Close()
				mu.Lock()
				service := portToService(p)
				gateways = append(gateways, GatewayInfo{
					Port:     p,
					Protocol: "tcp",
					Service:  service,
				})
				mu.Unlock()
			}
		}(port)
	}
	
	wg.Wait()
	return gateways
}

// GetAgents returns all discovered agents
func (d *DiscoveryService) GetAgents() []AgentInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]AgentInfo, 0, len(d.agents))
	for _, agent := range d.agents {
		result = append(result, *agent)
	}
	return result
}

// GetOnlineAgents returns only online agents
func (d *DiscoveryService) GetOnlineAgents() []AgentInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]AgentInfo, 0)
	for _, agent := range d.agents {
		if agent.Online {
			result = append(result, *agent)
		}
	}
	return result
}

// GetAgentByName returns a specific agent by name
func (d *DiscoveryService) GetAgentByName(name string) (*AgentInfo, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	agent, exists := d.agents[name]
	if !exists {
		return nil, false
	}
	
	// Return a copy
	agentCopy := *agent
	return &agentCopy, true
}

// GetAgentsJSON returns agents as JSON
func (d *DiscoveryService) GetAgentsJSON() (string, error) {
	agents := d.GetAgents()
	data, err := json.MarshalIndent(agents, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Helper functions
func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func portToService(port int) string {
	switch port {
	case 8001:
		return "bridge-inbound"
	case 8080:
		return "bridge-local"
	case 9090:
		return "agent-api"
	case 5000:
		return "http-alt"
	case 3000:
		return "dev-server"
	case 80:
		return "http"
	case 443:
		return "https"
	case 8443:
		return "https-alt"
	default:
		return "unknown"
	}
}

// Passive Discovery Methods (AIP-compatible - no port scanning)

func (d *DiscoveryService) runDiscoveryPassive() {
	for {
		select {
		case <-d.discoverTimer.C:
			d.discoverOncePassive()
		case <-d.stopChan:
			return
		}
	}
}

func (d *DiscoveryService) discoverOncePassive() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	status, err := d.localClient.Status(ctx)
	if err != nil {
		log.Printf("[discovery] failed to get status: %v", err)
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Mark all agents as offline first
	for _, agent := range d.agents {
		agent.Online = false
	}

	// Discover active peers from Tailscale (no port scanning)
	for _, peer := range status.Peer {
		if peer.DNSName == "" {
			continue
		}

		hostname := peer.DNSName
		if len(hostname) > 0 && hostname[len(hostname)-1] == '.' {
			hostname = hostname[:len(hostname)-1]
		}

		name := hostname
		if idx := indexOf(hostname, '.'); idx > 0 {
			name = hostname[:idx]
		}

		agent := d.agents[name]
		if agent == nil {
			agent = &AgentInfo{
				Name:     name,
				Hostname: hostname,
				Gateways: make([]GatewayInfo, 0),
			}
			d.agents[name] = agent
		}

		if len(peer.TailscaleIPs) > 0 {
			agent.IP = peer.TailscaleIPs[0].String()
		}
		agent.Online = true
		agent.LastSeen = time.Now()

		// NOTE: No port scanning in passive mode
		// Gateways remain empty - use AIP registration for capability discovery
		log.Printf("[discovery] passive: found peer %s (%s)", name, agent.IP)
	}

	// Remove stale agents (offline for > 5 minutes)
	for name, agent := range d.agents {
		if !agent.Online && time.Since(agent.LastSeen) > 5*time.Minute {
			delete(d.agents, name)
			log.Printf("[discovery] removed stale peer: %s", name)
		}
	}
}
