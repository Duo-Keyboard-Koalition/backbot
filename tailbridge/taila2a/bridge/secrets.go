package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	secretsFileName = "agent_secrets.json"
)

// AgentSecrets stores shared secrets for agent verification
type AgentSecrets struct {
	Secrets map[string]string `json:"secrets"` // agent_id -> secret
	mu      sync.RWMutex
}

// generateSecret creates a cryptographically secure random secret
func generateSecret() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// loadAgentSecrets loads agent secrets from the state directory
func loadAgentSecrets(stateDir string) map[string]string {
	secretsPath := filepath.Join(stateDir, secretsFileName)

	// Check if file exists
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		// Create empty secrets file
		secrets := &AgentSecrets{
			Secrets: make(map[string]string),
		}
		if err := saveAgentSecrets(secretsPath, secrets); err != nil {
			fmt.Printf("[secrets] warning: failed to create secrets file: %v\n", err)
		}
		return make(map[string]string)
	}

	// Read file
	data, err := os.ReadFile(secretsPath)
	if err != nil {
		fmt.Printf("[secrets] warning: failed to read secrets file: %v\n", err)
		return make(map[string]string)
	}

	// Parse JSON
	var secrets AgentSecrets
	if err := json.Unmarshal(data, &secrets); err != nil {
		fmt.Printf("[secrets] warning: failed to parse secrets file: %v\n", err)
		return make(map[string]string)
	}

	fmt.Printf("[secrets] loaded %d agent secrets\n", len(secrets.Secrets))
	return secrets.Secrets
}

// saveAgentSecrets saves agent secrets to file
func saveAgentSecrets(secretsPath string, secrets *AgentSecrets) error {
	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(secretsPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Write with restrictive permissions
	return os.WriteFile(secretsPath, data, 0600)
}

// AddAgentSecret adds a new agent secret
func AddAgentSecret(stateDir, agentID, secret string) error {
	secretsPath := filepath.Join(stateDir, secretsFileName)

	// Load existing secrets
	secrets := &AgentSecrets{
		Secrets: loadAgentSecrets(stateDir),
	}

	// Add/update secret
	secrets.mu.Lock()
	secrets.Secrets[agentID] = secret
	secrets.mu.Unlock()

	// Save
	return saveAgentSecrets(secretsPath, secrets)
}

// RemoveAgentSecret removes an agent secret
func RemoveAgentSecret(stateDir, agentID string) error {
	secretsPath := filepath.Join(stateDir, secretsFileName)

	// Load existing secrets
	secrets := &AgentSecrets{
		Secrets: loadAgentSecrets(stateDir),
	}

	// Remove secret
	secrets.mu.Lock()
	delete(secrets.Secrets, agentID)
	secrets.mu.Unlock()

	// Save
	return saveAgentSecrets(secretsPath, secrets)
}

// GetAgentSecret retrieves an agent secret
func GetAgentSecret(stateDir, agentID string) (string, bool) {
	secrets := loadAgentSecrets(stateDir)
	secret, exists := secrets[agentID]
	return secret, exists
}

// GenerateAgentSecret creates a new secret for an agent
func GenerateAgentSecret(stateDir, agentID string) (string, error) {
	secret, err := generateSecret()
	if err != nil {
		return "", err
	}

	if err := AddAgentSecret(stateDir, agentID, secret); err != nil {
		return "", err
	}

	return secret, nil
}

// ListAgentSecrets returns all agent IDs with secrets
func ListAgentSecrets(stateDir) []string {
	secrets := loadAgentSecrets(stateDir)
	agentIDs := make([]string, 0, len(secrets))
	for agentID := range secrets {
		agentIDs = append(agentIDs, agentID)
	}
	return agentIDs
}
