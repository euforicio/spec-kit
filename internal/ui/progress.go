package ui

import (
	"fmt"
	"time"
)

// ProgressTracker tracks progress of multi-step operations
type ProgressTracker struct {
	title   string
	steps   []Step
	active  bool
	current int
}

// Step represents a single step in a progress tracker
type Step struct {
	ID          string
	Description string
	Status      StepStatus
	Detail      string
	StartTime   time.Time
	EndTime     time.Time
}

// StepStatus represents the status of a step
type StepStatus int

const (
	StepStatusPending StepStatus = iota
	StepStatusRunning
	StepStatusCompleted
	StepStatusSkipped
	StepStatusFailed
)

// String returns a string representation of the step status
func (s StepStatus) String() string {
	switch s {
	case StepStatusPending:
		return "pending"
	case StepStatusRunning:
		return "running"
	case StepStatusCompleted:
		return "completed"
	case StepStatusSkipped:
		return "skipped"
	case StepStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(title string) *ProgressTracker {
	return &ProgressTracker{
		title: title,
		steps: make([]Step, 0),
	}
}

// AddStep adds a new step to the tracker
func (pt *ProgressTracker) AddStep(id, description string) {
	step := Step{
		ID:          id,
		Description: description,
		Status:      StepStatusPending,
	}
	pt.steps = append(pt.steps, step)
}

// Start begins progress tracking
func (pt *ProgressTracker) Start() {
	pt.active = true
	fmt.Printf("🚀 %s\n", pt.title)
}

// Stop ends progress tracking
func (pt *ProgressTracker) Stop() {
	pt.active = false
}

// StartStep marks a step as running
func (pt *ProgressTracker) StartStep(id string) {
	for i := range pt.steps {
		if pt.steps[i].ID == id {
			pt.steps[i].Status = StepStatusRunning
			pt.steps[i].StartTime = time.Now()
			pt.current = i
			pt.displayStep(&pt.steps[i])
			break
		}
	}
}

// CompleteStep marks a step as completed
func (pt *ProgressTracker) CompleteStep(id string) {
	for i := range pt.steps {
		if pt.steps[i].ID == id {
			pt.steps[i].Status = StepStatusCompleted
			pt.steps[i].EndTime = time.Now()
			pt.displayStep(&pt.steps[i])
			break
		}
	}
}

// SkipStep marks a step as skipped
func (pt *ProgressTracker) SkipStep(id, reason string) {
	for i := range pt.steps {
		if pt.steps[i].ID == id {
			pt.steps[i].Status = StepStatusSkipped
			pt.steps[i].Detail = reason
			pt.steps[i].EndTime = time.Now()
			pt.displayStep(&pt.steps[i])
			break
		}
	}
}

// FailStep marks a step as failed
func (pt *ProgressTracker) FailStep(id, error string) {
	for i := range pt.steps {
		if pt.steps[i].ID == id {
			pt.steps[i].Status = StepStatusFailed
			pt.steps[i].Detail = error
			pt.steps[i].EndTime = time.Now()
			pt.displayStep(&pt.steps[i])
			break
		}
	}
}

// displayStep displays the current state of a step
func (pt *ProgressTracker) displayStep(step *Step) {
	if !pt.active {
		return
	}

	var symbol string

	switch step.Status {
	case StepStatusPending:
		symbol = "○"
	case StepStatusRunning:
		symbol = "●"
	case StepStatusCompleted:
		symbol = "✓"
	case StepStatusSkipped:
		symbol = "○"
	case StepStatusFailed:
		symbol = "✗"
	}

	fmt.Printf("   %s %s", symbol, step.Description)

	if step.Detail != "" {
		fmt.Printf(" (%s)", step.Detail)
	}

	fmt.Println()
}

// GetSummary returns a summary of the progress
func (pt *ProgressTracker) GetSummary() string {
	completed := 0
	failed := 0
	skipped := 0

	for _, step := range pt.steps {
		switch step.Status {
		case StepStatusCompleted:
			completed++
		case StepStatusFailed:
			failed++
		case StepStatusSkipped:
			skipped++
		}
	}

	total := len(pt.steps)

	if failed > 0 {
		return fmt.Sprintf("%d/%d completed, %d failed", completed, total, failed)
	}

	if skipped > 0 {
		return fmt.Sprintf("%d/%d completed, %d skipped", completed, total, skipped)
	}

	return fmt.Sprintf("%d/%d completed", completed, total)
}

// IsCompleted returns true if all steps are completed or skipped
func (pt *ProgressTracker) IsCompleted() bool {
	for _, step := range pt.steps {
		if step.Status == StepStatusPending || step.Status == StepStatusRunning {
			return false
		}
	}
	return true
}

// HasFailed returns true if any step has failed
func (pt *ProgressTracker) HasFailed() bool {
	for _, step := range pt.steps {
		if step.Status == StepStatusFailed {
			return true
		}
	}
	return false
}
