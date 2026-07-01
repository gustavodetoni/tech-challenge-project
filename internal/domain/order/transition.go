package order

import (
	"fmt"
	"time"
)

func IsAllowedTransition(from, to Status) bool {
	switch from {
	case StatusReceived:
		return to == StatusInDiagnosis || to == StatusWaitingApproval
	case StatusInDiagnosis:
		return to == StatusWaitingApproval
	case StatusWaitingApproval:
		return to == StatusInProgress || to == StatusInDiagnosis
	case StatusInProgress:
		return to == StatusFinished
	case StatusFinished:
		return to == StatusDelivered
	default:
		return false
	}
}

func (o *ServiceOrder) TransitionTo(to Status, now time.Time) error {
	if o == nil {
		return fmt.Errorf("service order is required")
	}
	if o.Status == to {
		return nil
	}
	if !IsAllowedTransition(o.Status, to) {
		return fmt.Errorf("invalid status transition %s -> %s", o.Status, to)
	}
	o.Status = to
	o.UpdatedAt = now
	switch to {
	case StatusInDiagnosis:
		o.ExecutionStart = nil
	case StatusInProgress:
		o.ExecutionStart = &now
	case StatusFinished:
		o.FinishedAt = &now
	case StatusDelivered:
		o.DeliveredAt = &now
	}
	return nil
}

func (o *ServiceOrder) StartDiagnosis(now time.Time) error {
	return o.TransitionTo(StatusInDiagnosis, now)
}

func (o *ServiceOrder) SendBudget(now time.Time) error {
	return o.TransitionTo(StatusWaitingApproval, now)
}

func (o *ServiceOrder) StartExecution(now time.Time) error {
	return o.TransitionTo(StatusInProgress, now)
}

func (o *ServiceOrder) Finish(now time.Time) error {
	return o.TransitionTo(StatusFinished, now)
}

func (o *ServiceOrder) Deliver(now time.Time) error {
	return o.TransitionTo(StatusDelivered, now)
}
