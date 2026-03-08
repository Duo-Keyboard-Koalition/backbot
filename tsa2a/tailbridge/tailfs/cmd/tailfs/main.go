package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"tailscale.com/tsnet"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/tail-agent-file-send/internal/models"
	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/tail-agent-file-send/internal/services"
)

const Version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		runTailFS()
		return
	}

	command := os.Args[1]

	switch command {
	case "init":
		runInit()
	case "run":
		runTailFS()
	case "send":
		if len(os.Args) < 4 {
			fmt.Println("Usage: tailfs send <file> <destination-agent>")
			os.Exit(1)
		}
		sendFile(os.Args[2], os.Args[3])
	case "list":
		listAgents()
	case "status":
		showStatus()
	case "version":
		fmt.Printf("tailfs version %s\n", Version)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runTailFS() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Tailscale server
	srv := &tsnet.Server{
		Hostname: cfg.NodeName,
		Dir:      cfg.StateDir,
		AuthKey:  cfg.AuthKey,
	}
	defer srv.Close()

	// Initialize file transfer service
	transferConfig := models.DefaultFileTransferConfig()
	transferSvc, err := services.NewFileTransferService(transferConfig)
	if err != nil {
		log.Fatalf("Failed to create transfer service: %v", err)
	}
	transferSvc.Start()

	// Start HTTP server for API
	mux := http.NewServeMux()
	mux.HandleFunc("/send", handleSend(transferSvc))
	mux.HandleFunc("/receive", handleReceive(transferSvc))
	mux.HandleFunc("/progress", handleProgress(transferSvc))
	mux.HandleFunc("/history", handleHistory(transferSvc))
	mux.HandleFunc("/agents", handleAgents(srv))

	httpSrv := &http.Server{
		Addr:         cfg.LocalListen,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("tailfs node %q listening on %s", cfg.NodeName, cfg.LocalListen)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("tailfs serve failed: %v", err)
	}
}

func runInit() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== TailFS Setup ===")
	fmt.Println()

	fmt.Print("Tailscale Auth Key: ")
	authKey, _ := reader.ReadString('\n')
	authKey = strings.TrimSpace(authKey)

	fmt.Print("Node Name (e.g., tailfs-alpha): ")
	nodeName, _ := reader.ReadString('\n')
	nodeName = strings.TrimSpace(nodeName)
	if nodeName == "" {
		nodeName = "tailfs-default"
	}

	fmt.Print("Local Listen Address (default: 127.0.0.1:8081): ")
	localListen, _ := reader.ReadString('\n')
	localListen = strings.TrimSpace(localListen)
	if localListen == "" {
		localListen = "127.0.0.1:8081"
	}

	fmt.Print("Download Directory (default: ~/Downloads/tailfs): ")
	downloadDir, _ := reader.ReadString('\n')
	downloadDir = strings.TrimSpace(downloadDir)
	if downloadDir == "" {
		downloadDir = "~/Downloads/tailfs"
	}

	// Save config
	cfg := Config{
		NodeName:    nodeName,
		AuthKey:     authKey,
		LocalListen: localListen,
		DownloadDir: downloadDir,
	}

	if err := saveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✓ Setup complete!")
	fmt.Println("Run 'tailfs run' to start the service")
}

func sendFile(filePath, destAgent string) {
	fmt.Printf("Sending %s to %s...\n", filePath, destAgent)
	// TODO: Implement CLI file sending
}

func listAgents() {
	fmt.Println("Discovering agents on tailnet...")
	// TODO: Implement agent listing
}

func showStatus() {
	fmt.Println("TailFS Status")
	fmt.Println("=============")
	// TODO: Implement status display
}

func printUsage() {
	fmt.Println("TailFS - Secure File Transfer over Tailscale")
	fmt.Printf("Version: %s\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  tailfs init              Initialize configuration")
	fmt.Println("  tailfs run               Start tailfs service")
	fmt.Println("  tailfs send <file> <dst> Send file to agent")
	fmt.Println("  tailfs list              List available agents")
	fmt.Println("  tailfs status            Show transfer status")
	fmt.Println("  tailfs version           Show version")
	fmt.Println("  tailfs help              Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tailfs init")
	fmt.Println("  tailfs run")
	fmt.Println("  tailfs send document.pdf tailfs-beta")
	fmt.Println("  tailfs list")
}

// Config holds tailfs configuration
type Config struct {
	NodeName    string `json:"node_name"`
	StateDir    string `json:"state_dir"`
	AuthKey     string `json:"auth_key"`
	LocalListen string `json:"local_listen"`
	DownloadDir string `json:"download_dir"`
}

func loadConfig() (Config, error) {
	// TODO: Load from ~/.tailfs/config.json
	return Config{
		NodeName:    "tailfs-default",
		StateDir:    "~/.tailfs/state",
		LocalListen: "127.0.0.1:8081",
		DownloadDir: "~/Downloads/tailfs",
	}, nil
}

func saveConfig(cfg Config) error {
	// TODO: Save to ~/.tailfs/config.json
	return nil
}

// HTTP Handlers (placeholders)
func handleSend(svc *services.FileTransferService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "message": "send endpoint"}`)
	}
}

func handleReceive(svc *services.FileTransferService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "message": "receive endpoint"}`)
	}
}

func handleProgress(svc *services.FileTransferService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "transfers": []}`)
	}
}

func handleHistory(svc *services.FileTransferService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "history": []}`)
	}
}

func handleAgents(srv *tsnet.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "agents": []}`)
	}
}
