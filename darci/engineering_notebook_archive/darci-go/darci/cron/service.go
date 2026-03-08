// Package cron provides scheduled task functionality for darci-go.
package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"log/slog"
)

// nowMs returns the current time in milliseconds.
func nowMs() int64 {
	return time.Now().UnixMilli()
}

// OnJobFunc is a callback function type for executing jobs.
type OnJobFunc func(ctx context.Context, job CronJob) (string, error)

// CronService manages and executes scheduled jobs.
type CronService struct {
	mu            sync.RWMutex
	storePath     string
	onJob         OnJobFunc
	store         *CronStore
	timerTask     *time.Timer
	running       bool
	cancelFunc    context.CancelFunc
	logger        *slog.Logger
}

// CronServiceConfig holds configuration for the cron service.
type CronServiceConfig struct {
	StorePath string
	OnJob     OnJobFunc
}

// NewCronService creates a new cron service.
func NewCronService(cfg CronServiceConfig) *CronService {
	if cfg.StorePath == "" {
		homeDir, _ := os.UserHomeDir()
		cfg.StorePath = filepath.Join(homeDir, ".darci-go", "cron_store.json")
	}

	return &CronService{
		storePath:  cfg.StorePath,
		onJob:      cfg.OnJob,
		store:      &CronStore{Version: 1},
		logger:     slog.Default(),
	}
}

// loadStore loads jobs from disk.
func (c *CronService) loadStore() error {
	if c.store != nil && len(c.store.Jobs) > 0 {
		return nil
	}

	data, err := os.ReadFile(c.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.store = &CronStore{Version: 1, Jobs: make([]CronJob, 0)}
			return nil
		}
		return fmt.Errorf("failed to read cron store: %w", err)
	}

	var store CronStore
	if err := json.Unmarshal(data, &store); err != nil {
		return fmt.Errorf("failed to parse cron store: %w", err)
	}

	c.store = &store
	return nil
}

// saveStore saves jobs to disk.
func (c *CronService) saveStore() error {
	if c.store == nil {
		return nil
	}

	dir := filepath.Dir(c.storePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(c.store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cron store: %w", err)
	}

	if err := os.WriteFile(c.storePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cron store: %w", err)
	}

	return nil
}

// computeNextRun calculates the next run time for a schedule.
func computeNextRun(schedule CronSchedule, nowMs int64) *int64 {
	switch schedule.Kind {
	case ScheduleKindAt:
		if schedule.AtMs != nil && *schedule.AtMs > nowMs {
			return schedule.AtMs
		}
		return nil

	case ScheduleKindEvery:
		if schedule.EveryMs == nil || *schedule.EveryMs <= 0 {
			return nil
		}
		next := nowMs + *schedule.EveryMs
		return &next

	case ScheduleKindCron:
		if schedule.Expr == nil {
			return nil
		}
		// Simple implementation - in production, use a cron parser library
		// For now, return a placeholder
		next := nowMs + int64(time.Hour.Milliseconds())
		return &next
	}

	return nil
}

// validateSchedule validates a schedule for adding.
func validateSchedule(schedule CronSchedule) error {
	if schedule.Tz != nil && schedule.Kind != ScheduleKindCron {
		return fmt.Errorf("tz can only be used with cron schedules")
	}
	return nil
}

// recomputeNextRuns recomputes next run times for all enabled jobs.
func (c *CronService) recomputeNextRuns() {
	if c.store == nil {
		return
	}

	now := nowMs()
	for i := range c.store.Jobs {
		if c.store.Jobs[i].Enabled {
			c.store.Jobs[i].State.NextRunAtMs = computeNextRun(c.store.Jobs[i].Schedule, now)
		}
	}
}

// getNextWakeMs returns the earliest next run time across all jobs.
func (c *CronService) getNextWakeMs() *int64 {
	if c.store == nil {
		return nil
	}

	var earliest *int64
	for _, job := range c.store.Jobs {
		if job.Enabled && job.State.NextRunAtMs != nil {
			if earliest == nil || *job.State.NextRunAtMs < *earliest {
				earliest = job.State.NextRunAtMs
			}
		}
	}

	return earliest
}

// armTimer schedules the next timer tick.
func (c *CronService) armTimer(ctx context.Context) {
	if c.timerTask != nil {
		c.timerTask.Stop()
	}

	nextWake := c.getNextWakeMs()
	if nextWake == nil || !c.running {
		return
	}

	delayMs := *nextWake - nowMs()
	if delayMs < 0 {
		delayMs = 0
	}

	delay := time.Duration(delayMs) * time.Millisecond

	c.timerTask = time.AfterFunc(delay, func() {
		if c.running {
			c.onTimer(ctx)
		}
	})
}

// onTimer handles timer tick - runs due jobs.
func (c *CronService) onTimer(ctx context.Context) {
	if c.store == nil {
		return
	}

	now := nowMs()
	dueJobs := make([]CronJob, 0)

	for _, job := range c.store.Jobs {
		if job.Enabled && job.State.NextRunAtMs != nil && now >= *job.State.NextRunAtMs {
			dueJobs = append(dueJobs, job)
		}
	}

	for _, job := range dueJobs {
		c.executeJob(ctx, job)
	}

	if err := c.saveStore(); err != nil {
		c.logger.Error("Failed to save cron store", "error", err)
	}

	c.armTimer(ctx)
}

// executeJob executes a single job.
func (c *CronService) executeJob(ctx context.Context, job CronJob) {
	startMs := nowMs()
	c.logger.Info("Cron: executing job", "name", job.Name, "id", job.ID)

	var jobErr error

	if c.onJob != nil {
		_, jobErr = c.onJob(ctx, job)
	}

	// Find and update the job in store
	for i := range c.store.Jobs {
		if c.store.Jobs[i].ID == job.ID {
			if jobErr != nil {
				status := JobStatusError
				errStr := jobErr.Error()
				c.store.Jobs[i].State.LastStatus = &status
				c.store.Jobs[i].State.LastError = &errStr
				c.logger.Error("Cron: job failed", "name", job.Name, "error", jobErr)
			} else {
				status := JobStatusOK
				c.store.Jobs[i].State.LastStatus = &status
				c.store.Jobs[i].State.LastError = nil
				c.logger.Info("Cron: job completed", "name", job.Name)
			}

			c.store.Jobs[i].State.LastRunAtMs = &startMs
			c.store.Jobs[i].UpdatedAtMs = nowMs()

			// Handle one-shot jobs
			if job.Schedule.Kind == ScheduleKindAt {
				if job.DeleteAfterRun {
					// Remove job from store
					c.store.Jobs = append(c.store.Jobs[:i], c.store.Jobs[i+1:]...)
				} else {
					enabled := false
					c.store.Jobs[i].Enabled = enabled
					c.store.Jobs[i].State.NextRunAtMs = nil
				}
			} else {
				// Compute next run
				c.store.Jobs[i].State.NextRunAtMs = computeNextRun(job.Schedule, nowMs())
			}
			break
		}
	}
}

// Start starts the cron service.
func (c *CronService) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.loadStore(); err != nil {
		return err
	}

	c.running = true
	c.recomputeNextRuns()

	if err := c.saveStore(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel

	c.armTimer(ctx)

	c.logger.Info("Cron service started", "jobs", len(c.store.Jobs))
	return nil
}

// Stop stops the cron service.
func (c *CronService) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.running = false
	if c.cancelFunc != nil {
		c.cancelFunc()
		c.cancelFunc = nil
	}

	if c.timerTask != nil {
		c.timerTask.Stop()
		c.timerTask = nil
	}

	c.logger.Info("Cron service stopped")
	return nil
}

// ListJobs returns all jobs.
func (c *CronService) ListJobs(includeDisabled bool) []CronJob {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.store == nil {
		return nil
	}

	jobs := make([]CronJob, 0)
	for _, job := range c.store.Jobs {
		if includeDisabled || job.Enabled {
			jobs = append(jobs, job)
		}
	}

	// Sort by next run time
	sortJobsByNextRun(jobs)
	return jobs
}

// AddJob adds a new job.
func (c *CronService) AddJob(name string, schedule CronSchedule, message string, deliver bool, channel, to *string, deleteAfterRun bool) (*CronJob, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := validateSchedule(schedule); err != nil {
		return nil, err
	}

	if c.store == nil {
		if err := c.loadStore(); err != nil {
			return nil, err
		}
	}

	now := nowMs()

	job := CronJob{
		ID:      generateJobID(),
		Name:    name,
		Enabled: true,
		Schedule: schedule,
		Payload: CronPayload{
			Kind:    PayloadKindAgentTurn,
			Message: message,
			Deliver: deliver,
			Channel: channel,
			To:      to,
		},
		State: CronJobState{
			NextRunAtMs: computeNextRun(schedule, now),
		},
		CreatedAtMs:    now,
		UpdatedAtMs:    now,
		DeleteAfterRun: deleteAfterRun,
	}

	c.store.Jobs = append(c.store.Jobs, job)

	if err := c.saveStore(); err != nil {
		return nil, err
	}

	c.logger.Info("Cron: added job", "name", name, "id", job.ID)
	return &job, nil
}

// RemoveJob removes a job by ID.
func (c *CronService) RemoveJob(jobID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.store == nil {
		return false
	}

	before := len(c.store.Jobs)
	for i, job := range c.store.Jobs {
		if job.ID == jobID {
			c.store.Jobs = append(c.store.Jobs[:i], c.store.Jobs[i+1:]...)
			break
		}
	}

	removed := len(c.store.Jobs) < before
	if removed {
		if err := c.saveStore(); err != nil {
			c.logger.Error("Failed to save after removing job", "error", err)
		}
		c.logger.Info("Cron: removed job", "id", jobID)
	}

	return removed
}

// EnableJob enables or disables a job.
func (c *CronService) EnableJob(jobID string, enabled bool) (*CronJob, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.store == nil {
		return nil, fmt.Errorf("store not loaded")
	}

	for i := range c.store.Jobs {
		if c.store.Jobs[i].ID == jobID {
			c.store.Jobs[i].Enabled = enabled
			c.store.Jobs[i].UpdatedAtMs = nowMs()

			if enabled {
				c.store.Jobs[i].State.NextRunAtMs = computeNextRun(c.store.Jobs[i].Schedule, nowMs())
			} else {
				c.store.Jobs[i].State.NextRunAtMs = nil
			}

			if err := c.saveStore(); err != nil {
				return nil, err
			}

			return &c.store.Jobs[i], nil
		}
	}

	return nil, fmt.Errorf("job not found: %s", jobID)
}

// RunJob manually runs a job.
func (c *CronService) RunJob(ctx context.Context, jobID string, force bool) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.store == nil {
		return false, fmt.Errorf("store not loaded")
	}

	for _, job := range c.store.Jobs {
		if job.ID == jobID {
			if !force && !job.Enabled {
				return false, nil
			}
			c.executeJob(ctx, job)
			if err := c.saveStore(); err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return false, fmt.Errorf("job not found: %s", jobID)
}

// Status returns the service status.
func (c *CronService) Status() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := map[string]interface{}{
		"enabled": c.running,
		"jobs":    0,
	}

	if c.store != nil {
		status["jobs"] = len(c.store.Jobs)
	}

	if nextWake := c.getNextWakeMs(); nextWake != nil {
		status["next_wake_at_ms"] = *nextWake
	}

	return status
}

// sortJobsByNextRun sorts jobs by their next run time.
func sortJobsByNextRun(jobs []CronJob) {
	for i := 0; i < len(jobs)-1; i++ {
		for j := i + 1; j < len(jobs); j++ {
			if shouldSwap(jobs[i], jobs[j]) {
				jobs[i], jobs[j] = jobs[j], jobs[i]
			}
		}
	}
}

func shouldSwap(a, b CronJob) bool {
	if a.State.NextRunAtMs == nil && b.State.NextRunAtMs == nil {
		return false
	}
	if a.State.NextRunAtMs == nil {
		return true
	}
	if b.State.NextRunAtMs == nil {
		return false
	}
	return *a.State.NextRunAtMs > *b.State.NextRunAtMs
}

// generateJobID generates a unique job ID.
func generateJobID() string {
	return fmt.Sprintf("%d", nowMs())
}
