package cron

import (
	"time"
)

// ScheduleKind represents the type of schedule.
type ScheduleKind string

const (
	ScheduleKindAt     ScheduleKind = "at"
	ScheduleKindEvery  ScheduleKind = "every"
	ScheduleKindCron   ScheduleKind = "cron"
)

// JobStatus represents the status of a job execution.
type JobStatus string

const (
	JobStatusOK      JobStatus = "ok"
	JobStatusError   JobStatus = "error"
	JobStatusSkipped JobStatus = "skipped"
)

// PayloadKind represents the type of payload.
type PayloadKind string

const (
	PayloadKindSystemEvent PayloadKind = "system_event"
	PayloadKindAgentTurn   PayloadKind = "agent_turn"
)

// CronSchedule represents a schedule definition for a cron job.
type CronSchedule struct {
	Kind      ScheduleKind `json:"kind"`
	AtMs      *int64       `json:"at_ms,omitempty"`    // For "at": timestamp in ms
	EveryMs   *int64       `json:"every_ms,omitempty"` // For "every": interval in ms
	Expr      *string      `json:"expr,omitempty"`     // For "cron": cron expression
	Tz        *string      `json:"tz,omitempty"`       // Timezone for cron expressions
}

// CronPayload represents what to do when the job runs.
type CronPayload struct {
	Kind    PayloadKind `json:"kind"`
	Message string      `json:"message"`
	Deliver bool        `json:"deliver"`
	Channel *string     `json:"channel,omitempty"`
	To      *string     `json:"to,omitempty"`
}

// CronJobState represents the runtime state of a job.
type CronJobState struct {
	NextRunAtMs  *int64     `json:"next_run_at_ms,omitempty"`
	LastRunAtMs  *int64     `json:"last_run_at_ms,omitempty"`
	LastStatus   *JobStatus `json:"last_status,omitempty"`
	LastError    *string    `json:"last_error,omitempty"`
}

// CronJob represents a scheduled job.
type CronJob struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Enabled        bool          `json:"enabled"`
	Schedule       CronSchedule  `json:"schedule"`
	Payload        CronPayload   `json:"payload"`
	State          CronJobState  `json:"state"`
	CreatedAtMs    int64         `json:"created_at_ms"`
	UpdatedAtMs    int64         `json:"updated_at_ms"`
	DeleteAfterRun bool          `json:"delete_after_run"`
}

// CronStore represents the persistent store for cron jobs.
type CronStore struct {
	Version int        `json:"version"`
	Jobs    []CronJob  `json:"jobs"`
}

// CronJobResult holds the result of a job execution.
type CronJobResult struct {
	JobID     string    `json:"job_id"`
	Success   bool      `json:"success"`
	Output    string    `json:"output,omitempty"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

