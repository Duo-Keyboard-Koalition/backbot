package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/buffer"
	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/protocol"
)

// bufferSvc is the global buffer service instance
var bufferSvc *buffer.BufferService

// makeInboundHandler creates the handler for inbound messages from peer bridges
func makeInboundHandler(bridgeName, localAgentURL string, localClient *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		defer r.Body.Close()

		var env protocol.Envelope
		if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
			http.Error(w, "invalid envelope json", http.StatusBadRequest)
			return
		}
		if err := validateInboundEnvelope(&env); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Forward to local agent
		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, localAgentURL, bytes.NewReader(env.Payload))
		if err != nil {
			http.Error(w, "failed to create local agent request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := localClient.Do(req)
		if err != nil {
			http.Error(w, "local agent unreachable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("bridge copy response body error: %v", err)
		}

		log.Printf("[bridge:%s] delivered inbound payload from %s (%d)", bridgeName, env.SourceNode, resp.StatusCode)
	}
}

// makeSendHandler creates the handler for outbound messages with buffering
func makeSendHandler(bridgeName string, peerInboundPort int, tailnetClient *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		defer r.Body.Close()

		var env protocol.Envelope
		if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
			http.Error(w, "invalid envelope json", http.StatusBadRequest)
			return
		}
		if err := normalizeOutboundEnvelope(&env, bridgeName); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Create message for buffering
		msg := &buffer.Message{
			Envelope:   env,
			MaxRetries: 5,
		}

		// Attempt immediate delivery
		deliveryErr := attemptDelivery(r.Context(), msg, bridgeName, peerInboundPort, tailnetClient)
		if deliveryErr == nil {
			// Success - log and return response from destination
			log.Printf("[bridge:%s] routed outbound to %s (%d)", bridgeName, env.DestNode, http.StatusOK)
			return
		}

		// Delivery failed - buffer the message for retry
		log.Printf("[bridge:%s] delivery to %s failed: %v - buffering for retry", bridgeName, env.DestNode, deliveryErr)

		if bufferSvc == nil {
			http.Error(w, "destination bridge unreachable and buffer unavailable", http.StatusBadGateway)
			return
		}

		// Enqueue for buffered delivery
		bufferedMsg, err := bufferSvc.Enqueue(env, "")
		if err != nil {
			log.Printf("[bridge:%s] failed to buffer message: %v", bridgeName, err)
			http.Error(w, "failed to buffer message", http.StatusInternalServerError)
			return
		}

		// Attempt immediate delivery through buffer
		if err := bufferSvc.DeliverImmediately(bufferedMsg); err == nil {
			log.Printf("[bridge:%s] buffered message %s delivered successfully", bridgeName, bufferedMsg.ID)
			return
		}

		// Return accepted status - message will be retried in background
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		response := map[string]interface{}{
			"status":     "buffered",
			"message_id": bufferedMsg.ID,
			"destination": env.DestNode,
			"retry_count": bufferedMsg.RetryCount,
			"next_retry":  bufferedMsg.NextRetryAt,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("[bridge] failed to encode response: %v", err)
		}
	}
}

// attemptDelivery tries to deliver a message to the destination bridge
func attemptDelivery(ctx context.Context, msg *buffer.Message, bridgeName string, peerInboundPort int, tailnetClient *http.Client) error {
	payload, err := json.Marshal(msg.Envelope)
	if err != nil {
		return fmt.Errorf("failed to encode envelope: %w", err)
	}

	targetURL := fmt.Sprintf("http://%s:%d/inbound", msg.Envelope.DestNode, peerInboundPort)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create destination request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := tailnetClient.Do(req)
	if err != nil {
		return fmt.Errorf("destination bridge unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("destination returned status %d", resp.StatusCode)
	}

	return nil
}

// makeBufferDeliveryFunc creates the delivery function for the buffer service
func makeBufferDeliveryFunc(bridgeName string, peerInboundPort int, tailnetClient *http.Client) buffer.DeliveryFunc {
	return func(ctx context.Context, msg *buffer.Message) error {
		return attemptDelivery(ctx, msg, bridgeName, peerInboundPort, tailnetClient)
	}
}

// makeAgentsHandler creates the handler for listing discovered agents
func makeAgentsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if discoverySvc == nil {
			http.Error(w, "discovery service not initialized", http.StatusInternalServerError)
			return
		}

		agents := discoverySvc.GetOnlineAgents()
		w.Header().Set("Content-Type", "application/json")

		response := map[string]interface{}{
			"agents": agents,
			"count":  len(agents),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("[agents] failed to encode response: %v", err)
		}
	}
}

// makeBufferStatsHandler creates the handler for buffer statistics
func makeBufferStatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if bufferSvc == nil {
			http.Error(w, "buffer service not initialized", http.StatusInternalServerError)
			return
		}

		stats := bufferSvc.GetStats()
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(stats); err != nil {
			log.Printf("[buffer-stats] failed to encode response: %v", err)
		}
	}
}

// makeBufferMessagesHandler creates the handler for listing buffered messages
func makeBufferMessagesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if bufferSvc == nil {
			http.Error(w, "buffer service not initialized", http.StatusInternalServerError)
			return
		}

		status := r.URL.Query().Get("status")
		var messages []*buffer.Message
		var err error

		switch status {
		case "pending":
			messages, err = bufferSvc.GetPendingMessages()
		case "failed":
			messages, err = bufferSvc.GetFailedMessages()
		case "all":
			messages, err = bufferSvc.GetAllMessages()
		default:
			// Default to pending and retrying
			pending, _ := bufferSvc.GetPendingMessages()
			retrying, _ := bufferSvc.GetFailedMessages()
			messages = append(pending, retrying...)
		}

		if err != nil {
			http.Error(w, "failed to retrieve messages", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"messages": messages,
			"count":    len(messages),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("[buffer-messages] failed to encode response: %v", err)
		}
	}
}

// makeBufferRetryHandler creates the handler for retrying failed messages
func makeBufferRetryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if bufferSvc == nil {
			http.Error(w, "buffer service not initialized", http.StatusInternalServerError)
			return
		}

		var req struct {
			MessageID string `json:"message_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.MessageID == "" {
			http.Error(w, "message_id is required", http.StatusBadRequest)
			return
		}

		if err := bufferSvc.RetryFailedMessage(req.MessageID); err != nil {
			http.Error(w, fmt.Sprintf("failed to retry message: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"status":     "retrying",
			"message_id": req.MessageID,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("[buffer-retry] failed to encode response: %v", err)
		}
	}
}

// makeBufferClearHandler creates the handler for clearing delivered messages
func makeBufferClearHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if bufferSvc == nil {
			http.Error(w, "buffer service not initialized", http.StatusInternalServerError)
			return
		}

		count, err := bufferSvc.ClearDelivered()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to clear messages: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"cleared": count,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("[buffer-clear] failed to encode response: %v", err)
		}
	}
}

func runOutboundServer(localListen, bridgeName string, peerInboundPort int, tailnetClient *http.Client) {
	mux := http.NewServeMux()
	mux.HandleFunc("/send", makeSendHandler(bridgeName, peerInboundPort, tailnetClient))

	s := &http.Server{
		Addr:         localListen,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("bridge local server listening on %s (agent -> /send)", localListen)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("bridge local server failed: %v", err)
	}
}
