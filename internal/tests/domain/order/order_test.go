package order

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
)

func TestOrder_StatusValues_AreStable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "RECEIVED", string(order.StatusReceived))
	assert.Equal(t, "IN_DIAGNOSIS", string(order.StatusInDiagnosis))
	assert.Equal(t, "WAITING_APPROVAL", string(order.StatusWaitingApproval))
	assert.Equal(t, "IN_PROGRESS", string(order.StatusInProgress))
	assert.Equal(t, "FINISHED", string(order.StatusFinished))
	assert.Equal(t, "DELIVERED", string(order.StatusDelivered))
	assert.Equal(t, "CANCELED", string(order.StatusCanceled))

	values := []order.Status{
		order.StatusReceived,
		order.StatusInDiagnosis,
		order.StatusWaitingApproval,
		order.StatusInProgress,
		order.StatusFinished,
		order.StatusDelivered,
		order.StatusCanceled,
	}

	seen := map[string]struct{}{}
	for _, v := range values {
		_, exists := seen[string(v)]
		require.False(t, exists, "duplicate Status: %q", v)
		seen[string(v)] = struct{}{}
	}
}

func TestOrder_IsAllowedTransition(t *testing.T) {
	t.Parallel()

	assert.True(t, order.IsAllowedTransition(order.StatusReceived, order.StatusInDiagnosis))
	assert.True(t, order.IsAllowedTransition(order.StatusReceived, order.StatusWaitingApproval))
	assert.True(t, order.IsAllowedTransition(order.StatusInDiagnosis, order.StatusWaitingApproval))
	assert.True(t, order.IsAllowedTransition(order.StatusWaitingApproval, order.StatusInProgress))
	assert.True(t, order.IsAllowedTransition(order.StatusWaitingApproval, order.StatusInDiagnosis))
	assert.True(t, order.IsAllowedTransition(order.StatusInProgress, order.StatusFinished))
	assert.True(t, order.IsAllowedTransition(order.StatusFinished, order.StatusDelivered))

	assert.False(t, order.IsAllowedTransition(order.StatusReceived, order.StatusFinished))
	assert.False(t, order.IsAllowedTransition(order.StatusDelivered, order.StatusInDiagnosis))
}

func TestOrder_ServiceOrderTransitions_UpdateStatusAndTimestamps(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	serviceOrder := order.ServiceOrder{Status: order.StatusReceived}

	require.NoError(t, serviceOrder.StartDiagnosis(now))
	assert.Equal(t, order.StatusInDiagnosis, serviceOrder.Status)
	assert.Nil(t, serviceOrder.ExecutionStart)

	require.NoError(t, serviceOrder.SendBudget(now))
	assert.Equal(t, order.StatusWaitingApproval, serviceOrder.Status)

	require.NoError(t, serviceOrder.StartExecution(now))
	assert.Equal(t, order.StatusInProgress, serviceOrder.Status)
	require.NotNil(t, serviceOrder.ExecutionStart)
	assert.Equal(t, now, *serviceOrder.ExecutionStart)

	require.NoError(t, serviceOrder.Finish(now))
	assert.Equal(t, order.StatusFinished, serviceOrder.Status)
	require.NotNil(t, serviceOrder.FinishedAt)
	assert.Equal(t, now, *serviceOrder.FinishedAt)

	require.NoError(t, serviceOrder.Deliver(now))
	assert.Equal(t, order.StatusDelivered, serviceOrder.Status)
	require.NotNil(t, serviceOrder.DeliveredAt)
	assert.Equal(t, now, *serviceOrder.DeliveredAt)
}

func TestOrder_ServiceOrderTransitions_NoopsWhenStatusIsSame(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	serviceOrder := order.ServiceOrder{Status: order.StatusReceived}

	require.NoError(t, serviceOrder.TransitionTo(order.StatusReceived, now))
	assert.Equal(t, order.StatusReceived, serviceOrder.Status)
	assert.True(t, serviceOrder.UpdatedAt.IsZero())
}

func TestOrder_ServiceOrderTransitions_RejectInvalidTransition(t *testing.T) {
	t.Parallel()

	serviceOrder := order.ServiceOrder{Status: order.StatusReceived}

	err := serviceOrder.Finish(time.Now())

	require.Error(t, err)
	assert.Equal(t, order.StatusReceived, serviceOrder.Status)
}

func TestOrder_ServiceOrderTransitions_RequireServiceOrder(t *testing.T) {
	t.Parallel()

	var serviceOrder *order.ServiceOrder

	err := serviceOrder.TransitionTo(order.StatusInDiagnosis, time.Now())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "service order is required")
}

func TestOrder_BudgetTransitions(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	budget := order.Budget{Status: order.BudgetStatusDraft}

	require.NoError(t, budget.Send(now))
	assert.Equal(t, order.BudgetStatusSent, budget.Status)
	require.NotNil(t, budget.SentAt)

	name := "Maria"
	require.NoError(t, budget.Approve(now, &name))
	assert.Equal(t, order.BudgetStatusApproved, budget.Status)
	assert.Equal(t, &name, budget.ApprovedByName)
}

func TestOrder_BudgetRejectTransitions(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	budget := order.Budget{Status: order.BudgetStatusDraft}

	require.NoError(t, budget.Send(now))
	require.NoError(t, budget.Reject(now, "expensive"))

	assert.Equal(t, order.BudgetStatusRejected, budget.Status)
	require.NotNil(t, budget.RejectedAt)
	assert.Equal(t, now, *budget.RejectedAt)
	require.NotNil(t, budget.RejectionReason)
	assert.Equal(t, "expensive", *budget.RejectionReason)
}

func TestOrder_BudgetTransitions_RejectInvalidStatus(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)

	sendErr := (&order.Budget{Status: order.BudgetStatusApproved}).Send(now)
	approveErr := (&order.Budget{Status: order.BudgetStatusDraft}).Approve(now, nil)
	rejectErr := (&order.Budget{Status: order.BudgetStatusDraft}).Reject(now, "no")

	require.Error(t, sendErr)
	require.Error(t, approveErr)
	require.Error(t, rejectErr)
}

func TestOrder_BudgetTransitions_IgnoreNilBudget(t *testing.T) {
	t.Parallel()

	var budget *order.Budget
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)

	require.NoError(t, budget.Send(now))
	require.NoError(t, budget.Approve(now, nil))
	require.NoError(t, budget.Reject(now, "no"))
}

func TestOrder_BudgetStatusValues_AreStable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "DRAFT", string(order.BudgetStatusDraft))
	assert.Equal(t, "SENT", string(order.BudgetStatusSent))
	assert.Equal(t, "APPROVED", string(order.BudgetStatusApproved))
	assert.Equal(t, "REJECTED", string(order.BudgetStatusRejected))

	values := []order.BudgetStatus{
		order.BudgetStatusDraft,
		order.BudgetStatusSent,
		order.BudgetStatusApproved,
		order.BudgetStatusRejected,
	}

	seen := map[string]struct{}{}
	for _, v := range values {
		_, exists := seen[string(v)]
		require.False(t, exists, "duplicate BudgetStatus: %q", v)
		seen[string(v)] = struct{}{}
	}
}
