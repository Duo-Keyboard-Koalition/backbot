package models

// BufferServiceAdapter adapts the buffer service to the BufferMonitor interface
type BufferServiceAdapter struct {
	getPendingCountFunc func() (int, error)
	isRunningFunc       func() bool
}

// NewBufferServiceAdapter creates a new buffer service adapter
func NewBufferServiceAdapter(
	getPendingCountFunc func() (int, error),
	isRunningFunc func() bool,
) *BufferServiceAdapter {
	return &BufferServiceAdapter{
		getPendingCountFunc: getPendingCountFunc,
		isRunningFunc:       isRunningFunc,
	}
}

// GetPendingCount returns the number of pending messages
func (a *BufferServiceAdapter) GetPendingCount() (int, error) {
	if a.getPendingCountFunc != nil {
		return a.getPendingCountFunc()
	}
	return 0, nil
}

// IsRunning returns whether the buffer service is running
func (a *BufferServiceAdapter) IsRunning() bool {
	if a.isRunningFunc != nil {
		return a.isRunningFunc()
	}
	return false
}
