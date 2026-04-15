package admin_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestServiceService_Create_ValidatesAndSetsActive(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceService(repo)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(s *service.Service) bool {
		return s != nil &&
			s.ID != "" &&
			s.Name == "Troca de óleo" &&
			s.Active
	})).Return(nil).Once()

	_, err := svc.Create(context.Background(), service.Service{
		Name:             " Troca de óleo ",
		BasePriceCents:   10000,
		EstimatedMinutes: 30,
	})
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceService_Create_InvalidName(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceService(repo)

	_, err := svc.Create(context.Background(), service.Service{
		Name:             "   ",
		BasePriceCents:   10000,
		EstimatedMinutes: 30,
	})
	require.Error(t, err)
}
