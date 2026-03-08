package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"tailscale.com/tsnet"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/internal/controllers"
	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/internal/models"
	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/internal/views"
	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/cmd/taila2a/tui"
)

const Version = "0.2.0"

var discoverySvc *models.DiscoveryService

// Global agent state for trigger service
var (
	agentRunning      atomic.Bool
	agentCtx          context.Context
	agentCancel       context.CancelFunc
	bufferPendingCount atomic.Int32
)

func main() {
	if len(os.Args) < 2 {
		runAgnes()
		return
	}

	command := os.Args[1]

	switch command {
	case "init":
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "run":
		runAgnes()
	case "tui":
		if err := runTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Printf("agnes version %s\n", Version)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runTUI() error {
	cfg, err := models.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	port := cfg.PeerInboundPort
	if port == 0 {
		port = 8080 // Default fallback port
	}

	fmt.Printf("Starting SentinelAI TUI connected to localhost:%d\n", port)
	return tui.Run(port)
}

func runAgnes() {
	cfg, err := models.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Tailscale server
	srv := &tsnet.Server{
		Hostname: cfg.Name,
		Dir:      cfg.StateDir,
		AuthKey:  cfg.AuthKey,
	}
	defer srv.Close()

	// Initialize discovery service (Model)
	discoverySvc, err = models.NewDiscoveryService(srv)
	if err != nil {
		log.Fatalf("Failed to create discovery service: %v", err)
	}
	discoverySvc.Start(30 * time.Second)

	// Initialize HTTP clients
	tailnetClient := tsHTTPClient(srv, 20*time.Second)
	localClient := &http.Client{Timeout: 20 * time.Second}

	// Initialize TUI notifier
	tuiNotifier := models.NewConsoleTUINotifier(nil)

	// Initialize buffer adapter (simulated for now)
	bufferAdapter := models.NewBufferServiceAdapter(
		func() (int, error) {
			return int(bufferPendingCount.Load()), nil
		},
		func() bool {
			return true // Buffer service is always running in this simple impl
		},
	)

	// Initialize agent trigger service
	triggerSvc, err := models.NewAgentTriggerService(
		bufferAdapter,
		tuiNotifier,
		runAgent,      // startAgentFunc
		stopAgent,     // stopAgentFunc
		nil,           // use default config
	)
	if err != nil {
		log.Fatalf("Failed to create trigger service: %v", err)
	}
	triggerSvc.Start()

	// Initialize controller (Controller)
	controller := controllers.NewTaila2aController(
		cfg.Name,
		cfg.LocalAgentURL,
		cfg.PeerInboundPort,
		discoverySvc,
		tailnetClient,
		localClient,
	)

	// Initialize trigger controller
	triggerController := controllers.NewTriggerController(triggerSvc, tuiNotifier)

	// Initialize views (View)
	jsonView := views.NewJSONView()
	agentListView := views.NewAgentListView()

	// Start outbound server
	go runOutboundServer(cfg.LocalListen, cfg.Name, cfg.PeerInboundPort, tailnetClient, controller, triggerController)
	go logSelfTailscaleIPs(srv)

	// Start inbound server on tailnet
	ln, err := srv.Listen("tcp", fmt.Sprintf(":%d", cfg.InboundPort))
	if err != nil {
		log.Fatalf("agnes tailnet listen failed: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/inbound", controller.HandleInbound)
	mux.HandleFunc("/agents", controller.HandleAgents)
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		statusView := views.NewStatusView()
		status := map[string]interface{}{
			"name":     cfg.Name,
			"hostname": cfg.Name,
			"online":   true,
		}
		statusView.Render(w, status)
	})
	
	// Trigger endpoints
	mux.HandleFunc("/trigger/status", triggerController.HandleTriggerStatus)
	mux.HandleFunc("/trigger/manual", triggerController.HandleTriggerManual)
	mux.HandleFunc("/trigger/stop", triggerController.HandleTriggerStop)
	mux.HandleFunc("/trigger/notifications", triggerController.HandleNotifications)
	mux.HandleFunc("/trigger/notifications/clear", triggerController.HandleClearNotifications)

	httpSrv := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("taila2a node %q listening on tailnet :%d (state=%s)", cfg.Name, cfg.InboundPort, cfg.StateDir)
	log.Printf("discovery service started - agents available at /agents endpoint")
	log.Printf("agent trigger service started - buffer monitoring active")
	if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatalf("taila2a serve failed: %v", err)
	}

	// Keep references to avoid unused variable warnings
	_ = jsonView
	_ = agentListView
}

func runOutboundServer(localListen, name string, peerInboundPort int, tailnetClient *http.Client, controller *controllers.Taila2aController, triggerController *controllers.TriggerController) {
	mux := http.NewServeMux()
	mux.HandleFunc("/send", controller.HandleSend)
	
	// Endpoint to simulate buffer increase (for testing)
	mux.HandleFunc("/buffer/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		bufferPendingCount.Add(1)
		log.Printf("[buffer] pending count: %d", bufferPendingCount.Load())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true, "pending": %d}`, bufferPendingCount.Load())
	})
	
	// Endpoint to simulate buffer decrease (for testing)
	mux.HandleFunc("/buffer/remove", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		current := bufferPendingCount.Load()
		if current > 0 {
			bufferPendingCount.Add(-1)
		}
		log.Printf("[buffer] pending count: %d", bufferPendingCount.Load())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true, "pending": %d}`, bufferPendingCount.Load())
	})

	s := &http.Server{
		Addr:         localListen,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("taila2a local server listening on %s (agent -> /send)", localListen)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("taila2a local server failed: %v", err)
	}
}

// runAgent is called when the agent should start processing
// This is a placeholder - replace with actual agent logic
func runAgent(ctx context.Context) error {
	log.Printf("[agent] starting agent processing loop")
	
	// Simulate agent processing
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("[agent] agent context cancelled, stopping")
			return ctx.Err()
		case <-ticker.C:
			// Simulate processing a message from buffer
			current := bufferPendingCount.Load()
			if current > 0 {
				bufferPendingCount.Add(-1)
				log.Printf("[agent] processed message, remaining: %d", bufferPendingCount.Load())
			} else {
				log.Printf("[agent] buffer empty, agent stopping")
				return nil
			}
		}
	}
}

// stopAgent is called when the agent should stop
func stopAgent() error {
	log.Printf("[agent] stopAgent called")
	if agentCancel != nil {
		agentCancel()
	}
	agentRunning.Store(false)
	return nil
}

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
		log.Printf("agnes tailscale local client unavailable: %v", err)
		return
	}

	for i := 0; i < 30; i++ {
		st, err := lc.Status(context.Background())
		if err != nil {
			log.Printf("agnes tailscale status unavailable: %v", err)
			return
		}

		if st.Self != nil && len(st.Self.TailscaleIPs) > 0 {
			ips := make([]string, 0, len(st.Self.TailscaleIPs))
			for _, ip := range st.Self.TailscaleIPs {
				ips = append(ips, ip.String())
			}
			log.Printf("agnes tailnet IP(s): %s", strings.Join(ips, ", "))
			return
		}

		if i == 0 {
			log.Printf("taila2a tailnet IP pending: node not yet fully connected")
		}
		time.Sleep(2 * time.Second)
	}

	log.Printf("taila2a tailnet IP still pending after retry window")
}

func printUsage() {
	fmt.Println("Taila2a - Secure A2A Protocol over Tailscale")
	fmt.Printf("Version: %s\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  taila2a init    Initialize configuration interactively")
	fmt.Println("  taila2a run     Start taila2a (default)")
	fmt.Println("  taila2a tui     Start the Bubbletea dashboard")
	fmt.Println("  taila2a version Show version information")
	fmt.Println("  taila2a help    Show this help message")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Config file: ~/.taila2a/config.json")
	fmt.Println()
	fmt.Println("Endpoints:")
	fmt.Println("  POST /send                    - Send message to peer")
	fmt.Println("  GET  /agents                  - List discovered agents")
	fmt.Println("  GET  /status                  - Get taila2a status")
	fmt.Println("  GET  /trigger/status          - Get trigger service status")
	fmt.Println("  POST /trigger/manual          - Manually trigger agent")
	fmt.Println("  POST /trigger/stop            - Stop running agent")
	fmt.Println("  GET  /trigger/notifications   - Get TUI notifications")
	fmt.Println("  POST /trigger/notifications/clear - Clear notifications")
	fmt.Println("  POST /buffer/add              - Simulate buffer increase (test)")
	fmt.Println("  POST /buffer/remove           - Simulate buffer decrease (test)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  taila2a init                    # Interactive setup")
	fmt.Println("  taila2a run                     # Start taila2a with config")
	fmt.Println("  curl http://localhost:8080/trigger/status")
	fmt.Println("  curl -X POST http://localhost:8080/trigger/manual")
	fmt.Println("  curl -X POST http://localhost:8080/buffer/add")
}
