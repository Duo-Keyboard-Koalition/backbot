package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

// TestOrchestrator coordinates test execution across different environments
type TestOrchestrator struct {
	mode           string
	verbose        bool
	coverage       bool
	timeout        time.Duration
	agentCount     int
	cleanupOnExit  bool
}

func main() {
	orchestrator := &TestOrchestrator{}
	orchestrator.parseFlags()
	orchestrator.run()
}

func (o *TestOrchestrator) parseFlags() {
	flag.StringVar(&o.mode, "mode", "mock", "Test mode: mock, docker, or all")
	flag.StringVar(&o.mode, "m", "mock", "Shorthand for -mode")
	
	flag.BoolVar(&o.verbose, "v", false, "Verbose output")
	flag.BoolVar(&o.verbose, "verbose", false, "Verbose output")
	
	flag.BoolVar(&o.coverage, "coverage", false, "Generate coverage report")
	
	flag.DurationVar(&o.timeout, "timeout", 10*time.Minute, "Test timeout")
	
	flag.IntVar(&o.agentCount, "agents", 3, "Number of agents for integration tests")
	
	flag.BoolVar(&o.cleanupOnExit, "cleanup", true, "Cleanup on exit")
	
	flag.Parse()
}

func (o *TestOrchestrator) run() {
	log.Printf("Starting Test Orchestrator")
	log.Printf("Mode: %s", o.mode)
	log.Printf("Verbose: %v", o.verbose)
	log.Printf("Coverage: %v", o.coverage)
	log.Printf("Timeout: %v", o.timeout)
	log.Printf("")

	startTime := time.Now()
	
	switch o.mode {
	case "mock":
		o.runMockTests()
	case "docker":
		o.runDockerTests()
	case "all":
		o.runMockTests()
		o.runDockerTests()
	default:
		log.Fatalf("Unknown mode: %s", o.mode)
	}
	
	elapsed := time.Since(startTime)
	log.Printf("")
	log.Printf("Test execution completed in %v", elapsed)
}

func (o *TestOrchestrator) runMockTests() {
	log.Println("========================================")
	log.Println("Running Mock Tests")
	log.Println("========================================")
	
	args := []string{"test", "./mock/..."}
	
	if o.verbose {
		args = append(args, "-v")
	}
	
	if o.coverage {
		args = append(args, "-coverprofile=mock-coverage.out")
	}
	
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "."
	
	if err := cmd.Run(); err != nil {
		log.Fatalf("Mock tests failed: %v", err)
	}
	
	log.Println("Mock tests PASSED")
	log.Println("")
}

func (o *TestOrchestrator) runDockerTests() {
	log.Println("========================================")
	log.Println("Running Docker Integration Tests")
	log.Println("========================================")
	
	// Check if docker-compose is available
	if _, err := exec.LookPath("docker-compose"); err != nil {
		if _, err := exec.LookPath("docker"); err != nil {
			log.Fatal("Docker is not available. Please install Docker and docker-compose.")
		}
		// Try docker compose (v2)
		o.runDockerComposeV2()
		return
	}
	
	o.runDockerComposeV1()
}

func (o *TestOrchestrator) runDockerComposeV1() {
	log.Println("Using docker-compose (v1)")
	
	// Start containers
	log.Println("Starting Docker containers...")
	cmd := exec.Command("docker-compose", "-f", "docker/docker-compose.test.yml", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to start Docker containers: %v", err)
	}
	
	defer o.cleanupDocker()
	
	// Wait for agents
	log.Println("Waiting for agents to be ready...")
	time.Sleep(30 * time.Second)
	
	// Run integration tests
	args := []string{"test", "./integration/...", "-tags=integration", "-timeout", o.timeout.String()}
	
	if o.verbose {
		args = append(args, "-v")
	}
	
	cmd = exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		log.Fatalf("Integration tests failed: %v", err)
	}
	
	log.Println("Integration tests PASSED")
	log.Println("")
}

func (o *TestOrchestrator) runDockerComposeV2() {
	log.Println("Using docker compose (v2)")
	
	// Start containers
	log.Println("Starting Docker containers...")
	cmd := exec.Command("docker", "compose", "-f", "docker/docker-compose.test.yml", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to start Docker containers: %v", err)
	}
	
	defer o.cleanupDockerV2()
	
	// Wait for agents
	log.Println("Waiting for agents to be ready...")
	time.Sleep(30 * time.Second)
	
	// Run integration tests
	args := []string{"test", "./integration/...", "-tags=integration", "-timeout", o.timeout.String()}
	
	if o.verbose {
		args = append(args, "-v")
	}
	
	cmd = exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		log.Fatalf("Integration tests failed: %v", err)
	}
	
	log.Println("Integration tests PASSED")
	log.Println("")
}

func (o *TestOrchestrator) cleanupDocker() {
	if !o.cleanupOnExit {
		return
	}
	
	log.Println("Cleaning up Docker containers...")
	cmd := exec.Command("docker-compose", "-f", "docker/docker-compose.test.yml", "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func (o *TestOrchestrator) cleanupDockerV2() {
	if !o.cleanupOnExit {
		return
	}
	
	log.Println("Cleaning up Docker containers...")
	cmd := exec.Command("docker", "compose", "-f", "docker/docker-compose.test.yml", "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func init() {
	// Verify we're in the right directory
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("Error: Please run from the test_platform directory")
		os.Exit(1)
	}
}
