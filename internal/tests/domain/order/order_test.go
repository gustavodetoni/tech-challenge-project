package order

import (
	"testing"

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
