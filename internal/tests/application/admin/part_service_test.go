package admin_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestPartService_Create_ValidatesAndSetsActive(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.PartRepository)
	svc := admin.NewPartInventoryUseCase(repo)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(p *part.Part) bool {
		return p != nil &&
			p.ID != "" &&
			p.SKU == "SKU1" &&
			p.Name == "Filtro" &&
			p.Active
	})).Return(nil).Once()

	_, err := svc.Create(context.Background(), part.Part{
		SKU:            " SKU1 ",
		Name:           " Filtro ",
		UnitPriceCents: 100,
		StockQuantity:  10,
	})
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestPartService_CreateFromInput_MapsInput(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.PartRepository)
	svc := admin.NewPartInventoryUseCase(repo)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(p *part.Part) bool {
		return p != nil &&
			p.SKU == "SKU1" &&
			p.Name == "Filtro" &&
			p.UnitPriceCents == 100 &&
			p.StockQuantity == 10
	})).Return(nil).Once()

	_, err := svc.CreateFromInput(context.Background(), admin.CreatePartInput{
		SKU:            "SKU1",
		Name:           "Filtro",
		UnitPriceCents: 100,
		StockQuantity:  10,
	})
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestPartService_Create_InvalidStock(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.PartRepository)
	svc := admin.NewPartInventoryUseCase(repo)

	_, err := svc.Create(context.Background(), part.Part{
		SKU:            "SKU1",
		Name:           "Filtro",
		UnitPriceCents: 100,
		StockQuantity:  -1,
	})
	require.Error(t, err)
}
