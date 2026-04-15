package admin_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestServiceOrderService_FindByID_CallsRepo(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderService(repo)

	repo.On("FindByID", mock.Anything, "so1").Return(&order.ServiceOrder{ID: "so1"}, nil).Once()

	out, err := svc.FindByID(context.Background(), "so1")
	require.NoError(t, err)
	require.Equal(t, "so1", out.ID)
	repo.AssertExpectations(t)
}
